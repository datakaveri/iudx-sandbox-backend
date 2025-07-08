package restartnotebook

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/cmd/tasks/handlers/restartnotebook"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/jupyterutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func restartNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Create context with request ID for tracing
		ctx := context.WithValue(r.Context(), "request_id", uuid.New().String())

		logger.InfoWithContext(ctx, "Starting notebook restart request")

		defer r.Body.Close()

		tokenUser, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.ErrorWithMetadata("Authentication failed during notebook restart", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Add user ID to context for logging
		ctx = context.WithValue(ctx, "user_id", tokenUser.UserName)

		logger.DebugWithContext(ctx, "Authentication successful for notebook restart")

		userModel := &models.User{}
		user, err := userModel.Get(app, tokenUser.UserName)
		if err != nil {
			logger.ErrorWithMetadata("User lookup failed during notebook restart", map[string]interface{}{
				"username": tokenUser.UserName,
			}, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.DebugWithMetadata("User details retrieved for restart", map[string]interface{}{
			"user_id":  user.UserId,
			"username": user.Name,
		})

		queryValues := r.URL.Query()
		notebook := &models.Notebook{}
		notebookId := queryValues.Get("notebookId")
		notebook.NotebookId = notebookId

		logger.InfoWithMetadata("Notebook restart requested", map[string]interface{}{
			"notebook_id": notebookId,
			"user_id":     user.UserId,
		})

		jupyterClient, err := jupyterutility.Get()
		if err != nil {
			logger.ErrorWithMetadata("Jupyter client failure during restart", map[string]interface{}{
				"notebook_id": notebookId,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		notebookData, err := notebook.Get(app, notebook.NotebookId)
		if err != nil {
			logger.ErrorWithMetadata("No notebook found for restart", map[string]interface{}{
				"notebook_id": notebookId,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		spawnerName, err := notebook.GetSpawnerName(app, user.UserId, notebookId)
		if err != nil {
			logger.ErrorWithMetadata("Failed to get spawner name for restart", map[string]interface{}{
				"notebook_id": notebookId,
				"user_id":     user.UserId,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, endpoint, err := jupyterClient.RestartServer(app, tokenUser.UserName, spawnerName)
		if err != nil || endpoint == "" {
			logger.ErrorWithMetadata("Jupyter client failure during restart", map[string]interface{}{
				"notebook_id":  notebookId,
				"spawner_name": spawnerName,
				"endpoint":     endpoint,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.InfoWithMetadata("Jupyter restart endpoint obtained", map[string]interface{}{
			"notebook_id": notebookId,
			"build_id":    notebookData.BuildId,
			"endpoint":    endpoint,
		})

		progressUrl := fmt.Sprintf("%s/progress", endpoint)

		// Add restart notebook task to the queue
		if err := restartnotebook.AddRestartNotebookTask(ctx, app, progressUrl, notebookData.BuildId, user.UserId); err != nil {
			logger.ErrorWithMetadata("Failed to add restart notebook task to queue", map[string]interface{}{
				"build_id":     notebookData.BuildId,
				"user_id":      user.UserId,
				"progress_url": progressUrl,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.InfoWithMetadata("Restart notebook task added to queue successfully", map[string]interface{}{
			"build_id":  notebookData.BuildId,
			"task_type": "restart-notebook",
		})

		// Update notebook status to restarting
		notebook.Phase = "restarting"
		notebook.BuildId = notebookData.BuildId
		notebook.Message = "Notebook is restarting"

		// Use UpdateNotebookBuildStatus for consistency with build status endpoint
		err = notebook.UpdateNotebookBuildStatus(app)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.InfoWithMetadata("No records found for notebook restart", map[string]interface{}{
					"notebook_id": notebookId,
				})
				w.WriteHeader(http.StatusPreconditionFailed)
				return
			}

			logger.ErrorWithMetadata("Error updating notebook build status to restarting", map[string]interface{}{
				"notebook_id": notebookId,
				"build_id":    notebookData.BuildId,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Audit log for notebook restart
		logger.AuditLog(ctx, "notebook_restart_started", "notebook", map[string]interface{}{
			"notebook_id": notebookId,
			"build_id":    notebookData.BuildId,
		})

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook restarting. Please check the status")
		dataResponse := newResponse.AddData(map[string]string{
			"buildId": notebookData.BuildId,
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)

		logger.InfoWithContext(ctx, "Notebook restart request completed successfully")
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(restartNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
