package buildnotebook

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/r3labs/sse/v2"
)

func handleBinderSSE(app *application.Application, msg *sse.Event, buildId string) {
	buildlog := &models.BuildLog{}
	json.Unmarshal(msg.Data, buildlog)
	buildlog.BuildId = buildId
	if err := buildlog.Create(app); err != nil {
		logger.Error.Printf("Binder: Error in building notebook %v\n", err)
		return
	}

	if buildlog.Phase == "ready" {
		notebook := &models.Notebook{}

		if err := notebook.UpdateNotebookReady(app, buildId, "RUNNING", buildlog.NotebookUrl); err != nil {
			logger.Error.Printf("Binder: Error in building notebook %v\n", err)
			return
		}
	}
}

func buildNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		notebook := &models.Notebook{}
		json.NewDecoder(r.Body).Decode(notebook)
		notebook.BuildStatus = "Building"
		notebook.BuildId = uuid.New().String()

		sseClient := sse.NewClient(app.Cfg.GetBinderNotebookBuildApi("gh", notebook.RepoName))

		sseClient.SubscribeRaw(func(msg *sse.Event) {
			handleBinderSSE(app, msg, notebook.BuildId)
		})

		if err := notebook.Create(app); err != nil {
			logger.Error.Printf("Error in building notebook %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook Building. Please check the status")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(buildNotebook(app), middleware.LogRequest)
}
