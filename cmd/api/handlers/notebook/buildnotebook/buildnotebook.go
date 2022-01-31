package buildnotebook

import (
	"encoding/json"
	"net/http"

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

func handleBinderSSE(app *application.Application, msg *sse.Event, buildId string) {
	notebook := &models.Notebook{}
	json.Unmarshal(msg.Data, notebook)
	notebook.BuildId = buildId

	if err := notebook.UpdateNotebookStatus(app); err != nil {
		logger.Error.Printf("Binder: Error in building notebook %v\n", err)
		return
	}
}

func makeSseRequest(app *application.Application, buildUrl, cookie, buildId string) {
	sseClient := sse.NewClient(buildUrl, CustomHeader(cookie))

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		go handleBinderSSE(app, msg, buildId)
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

		go makeSseRequest(app, buildUrl, cookie, notebook.BuildId)

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
