package buildnotebook

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/r3labs/sse/v2"
)

func handleBinderSSE(app *application.Application, msg *sse.Event, buildId string, userId int) {
	notebook := &models.Notebook{}
	json.Unmarshal(msg.Data, notebook)
	notebook.BuildId = buildId

	if notebook.Phase == "ready" {
		// FIXME slightly inconsistent approach its unreliable maybe we can change spawner name but for single server that won't work
		parsedUrl, err := url.Parse(notebook.NotebookUrl)
		baseUrl := parsedUrl.Path + "/"
		if err != nil {
			logger.Error.Printf("Error parsing url %v\n", err)
		}
		spawner := &models.Spawner{}
		res, err := spawner.GetSpawnerIdBasedOnBaseUrl(app, baseUrl, userId)
		if err != nil {
			logger.Error.Printf("Error finding spawner id %v\n", err)
		}
		notebook.SpawnerId = res.Id
		if err := notebook.UpdateNotebookSpawnerId(app); err != nil {
			logger.Error.Printf("Error failed to update spawner id %v\n", err)
		}
	}

	if err := notebook.UpdateNotebookBuildStatus(app); err != nil {
		logger.Error.Printf("Binder: Error in building notebook %v\n", err)
		return
	}
}

func makeSseRequest(app *application.Application, buildUrl, cookie, buildId string, userId int) {
	sseClient := sse.NewClient(buildUrl, CustomHeader(cookie))

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		go handleBinderSSE(app, msg, buildId, userId)
	})

	if sseserror != nil {
		logger.Error.Printf("Error in sse %v\n", sseserror)
	}
}

func CustomHeader(cookie string) func(c *sse.Client) {
	return func(c *sse.Client) {
		c.Headers = map[string]string{
			"cookie": cookie,
		}
	}
}

func buildNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		defer r.Body.Close()
		tokenUser, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.Error.Printf("Error in building notebook, Unauthorized %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userModel := &models.User{}
		user, err := userModel.Get(app, tokenUser.UserName)

		if err != nil {
			logger.Error.Printf("Error in building notebook, User not found %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		notebook := &models.Notebook{}
		json.NewDecoder(r.Body).Decode(notebook)

		// check if the notebook server is already present
		notebookId, err := notebook.GetNotebookIdByRepoName(app, notebook.RepoName, user.UserId)

		if notebookId != "" {
			logger.Error.Printf("Error in building notebook, Notebook already exists %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			newResponse := apiresponse.New("error", "Notebook already exists. Please restart or use the same")
			response, _ := newResponse.Marshal()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(response)
			return
		}

		notebook.NotebookId = uuid.New().String()
		notebook.BuildId = uuid.New().String()
		notebook.UserId = user.UserId
		notebook.Phase = "building"

		// token := r.Header.Get("Authorization")
		// splitToken := strings.Split(token, "Bearer ")
		// token = splitToken[1]

		cookie := r.Header.Get("BuildToken")

		buildUrl := app.Cfg.GetBinderNotebookBuildApi(notebook.RepoName)

		if err := notebook.Create(app); err != nil {
			logger.Error.Printf("Error in building notebook %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go makeSseRequest(app, buildUrl, cookie, notebook.BuildId, notebook.UserId)

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook Building. Please check the status")
		dataResponse := newResponse.AddData(map[string]string{
			"buildId": notebook.BuildId,
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(buildNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
