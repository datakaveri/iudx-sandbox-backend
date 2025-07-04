package buildnotebook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/cmd/tasks/handlers/spawnernotebooksync"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func buildNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Create context with request ID for tracing
		ctx := context.WithValue(r.Context(), "request_id", uuid.New().String())

		logger.InfoWithContext(ctx, "Starting notebook build request")

		defer r.Body.Close()

		// Extract and validate authentication
		tokenUser, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.ErrorWithMetadata("Authentication failed during notebook build", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Add user ID to context for logging
		ctx = context.WithValue(ctx, "user_id", tokenUser.UserName)

		logger.DebugWithContext(ctx, "Authentication successful for notebook build")

		// Get user details
		userModel := &models.User{}
		user, err := userModel.Get(app, tokenUser.UserName)
		if err != nil {
			logger.ErrorWithMetadata("User lookup failed during notebook build", map[string]interface{}{
				"username": tokenUser.UserName,
			}, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.DebugWithMetadata("User details retrieved", map[string]interface{}{
			"user_id":  user.UserId,
			"username": user.Name,
		})

		// Parse notebook request
		notebook := &models.Notebook{}
		if err := json.NewDecoder(r.Body).Decode(notebook); err != nil {
			logger.WarnWithError("Invalid JSON in notebook build request", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger.InfoWithMetadata("Notebook build requested", map[string]interface{}{
			"repo_name": notebook.RepoName,
			"user_id":   user.UserId,
		})

		// Check for existing notebook
		existingNotebookId, err := notebook.GetNotebookIdByRepoName(app, notebook.RepoName, user.UserId)
		if existingNotebookId != "" {
			logger.WarnWithMetadata("Notebook already exists", map[string]interface{}{
				"repo_name":            notebook.RepoName,
				"existing_notebook_id": existingNotebookId,
				"user_id":              user.UserId,
			})

			w.Header().Set("Content-Type", "application/json")
			newResponse := apiresponse.New("error", "Notebook already exists. Please restart or use the same")
			response, _ := newResponse.Marshal()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(response)
			return
		}

		// Generate IDs and set notebook properties
		notebook.NotebookId = uuid.New().String()
		notebook.BuildId = uuid.New().String()
		notebook.UserId = user.UserId
		notebook.Phase = "building"

		cookie := r.Header.Get("BuildToken")
		buildUrl := app.Cfg.GetBinderNotebookBuildApi(notebook.RepoName)

		logger.InfoWithMetadata("Notebook configuration prepared", map[string]interface{}{
			"notebook_id":   notebook.NotebookId,
			"build_id":      notebook.BuildId,
			"build_url":     buildUrl,
			"repo_name":     notebook.RepoName,
			"has_cookie":    cookie != "",
			"cookie_length": len(cookie),
			"cookie_value":  cookie,
		})

		// Create notebook record
		if err := notebook.Create(app); err != nil {
			logger.ErrorWithMetadata("Failed to create notebook record", map[string]interface{}{
				"notebook_id": notebook.NotebookId,
				"build_id":    notebook.BuildId,
				"repo_name":   notebook.RepoName,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.InfoWithMetadata("Notebook record created successfully", map[string]interface{}{
			"notebook_id": notebook.NotebookId,
			"build_id":    notebook.BuildId,
		})

		// Add spawner notebook sync task to the queue
		if err := spawnernotebooksync.AddSpawnerSyncTask(ctx, app, buildUrl, cookie, notebook.BuildId, notebook.UserId); err != nil {
			logger.ErrorWithMetadata("Failed to add spawner sync task to queue", map[string]interface{}{
				"build_id":  notebook.BuildId,
				"user_id":   notebook.UserId,
				"build_url": buildUrl,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.InfoWithMetadata("Spawner sync task added to queue successfully", map[string]interface{}{
			"build_id":  notebook.BuildId,
			"task_type": "spawner-notebook-sync",
		})

		// Audit log for notebook creation
		logger.AuditLog(ctx, "notebook_build_started", "notebook", map[string]interface{}{
			"notebook_id": notebook.NotebookId,
			"build_id":    notebook.BuildId,
			"repo_name":   notebook.RepoName,
		})

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook Building. Please check the status")
		dataResponse := newResponse.AddData(map[string]string{
			"buildId": notebook.BuildId,
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)

		logger.InfoWithContext(ctx, "Notebook build request completed successfully")
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(buildNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
