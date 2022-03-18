package restartnotebook

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/jupyterutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/r3labs/sse/v2"
)

type RestartProgressResponse struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

func handleJupyterSSE(app *application.Application, msg *sse.Event, notebook *models.Notebook) {
	response := &RestartProgressResponse{}
	json.Unmarshal(msg.Data, response)
	fmt.Println(response)

	if response.Ready == true {
		// FIXME slightly inconsistent approach its unreliable maybe we can change spawner name but for single server that won't work
		spawner := &models.Spawner{}
		res, err := spawner.GetSpawnerIdBasedOnBaseUrl(app, response.Url, notebook.UserId)
		if err != nil {
			logger.Error.Printf("Error finding spawner id %v\n", err)
		}
		notebook.SpawnerId = res.Id
		if err := notebook.UpdateNotebookSpawnerId(app); err != nil {
			logger.Error.Printf("Error failed to update spawner id %v\n", err)
		}

		notebook.Phase = "ready"

		if err := notebook.UpdateNotebookStatus(app); err != nil {
			logger.Error.Printf("Binder: Error in restarting notebook %v\n", err)
			return
		}
		return
	}

	notebook.Phase = "restarting"

	if err := notebook.UpdateNotebookStatus(app); err != nil {
		logger.Error.Printf("Binder: Error in restarting notebook %v\n", err)
		return
	}
}

func makeSseRequest(app *application.Application, progressUrl string, notebook *models.Notebook) {
	fmt.Println("Making restart request", progressUrl)
	sseClient := sse.NewClient(progressUrl, CustomHeader(app))

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		go handleJupyterSSE(app, msg, notebook)
	})

	if sseserror != nil {
		logger.Error.Printf("Error in sse %v\n", sseserror)
	}
}

func CustomHeader(app *application.Application) func(c *sse.Client) {
	return func(c *sse.Client) {
		c.Headers = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", app.Cfg.GetJupyterHubApiToken()),
		}
	}
}

func restartNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		tokenUser, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.Error.Printf("Error in fetching notebook, Unauthorized %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userModel := &models.User{}
		user, err := userModel.Get(app, tokenUser.UserName)

		queryValues := r.URL.Query()
		notebook := &models.Notebook{}
		notebookId := queryValues.Get("notebookId")
		notebook.NotebookId = notebookId

		jupyterClient, err := jupyterutility.Get()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure")
			return
		}

		notebookData, err := notebook.Get(app, notebook.NotebookId)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("No notebook found", err)
			return
		}

		spawnerName, err := notebook.GetSpawnerName(app, user.UserId, notebookId)

		_, endpoint, err := jupyterClient.RestartServer(app, tokenUser.UserName, spawnerName)

		if err != nil || endpoint == "" {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure", err)
			return
		}

		go makeSseRequest(app, fmt.Sprintf("%s/progress", endpoint), &notebookData)

		notebook.Phase = "restarting"
		err = notebook.UpdateNotebookStatus(app)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error in restarting notebook %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook starting")
		dataResponse := newResponse.AddData(map[string]string{
			"buildId": notebook.BuildId,
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(restartNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
