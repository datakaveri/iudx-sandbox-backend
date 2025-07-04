package notebookbuildstatus

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func getBuildStatus(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := r.Context()
		start := time.Now()

		logger.InfoWithContext(ctx, "Starting notebook build status request")

		queryValues := r.URL.Query()
		buildId := queryValues.Get("buildId")

		if buildId == "" {
			logger.WarnWithMetadata("Build status request missing buildId parameter", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
			})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger.DebugWithMetadata("Checking build status", map[string]interface{}{
			"build_id":    buildId,
			"remote_addr": r.RemoteAddr,
		})

		notebook := &models.Notebook{}
		res, err := notebook.GetBuildStatus(app, buildId)
		queryDuration := time.Since(start)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.InfoWithMetadata("No build record found", map[string]interface{}{
					"build_id":       buildId,
					"remote_addr":    r.RemoteAddr,
					"query_duration": queryDuration.String(),
				})
				w.WriteHeader(http.StatusNoContent)
				return
			}

			logger.ErrorWithMetadata("Error fetching notebook build status", map[string]interface{}{
				"build_id":       buildId,
				"remote_addr":    r.RemoteAddr,
				"query_duration": queryDuration.String(),
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Log performance metrics
		if queryDuration > 1*time.Second {
			logger.WarnWithMetadata("Slow build status query detected", map[string]interface{}{
				"build_id":       buildId,
				"query_duration": queryDuration.String(),
				"remote_addr":    r.RemoteAddr,
			})
		}

		logger.InfoWithMetadata("Build status retrieved successfully", map[string]interface{}{
			"build_id":       buildId,
			"phase":          res.Phase,
			"remote_addr":    r.RemoteAddr,
			"query_duration": queryDuration.String(),
		})

		// Audit log for build status access
		logger.AuditLog(ctx, "notebook_build_status_accessed", "notebook", map[string]interface{}{
			"build_id":    buildId,
			"phase":       res.Phase,
			"remote_addr": r.RemoteAddr,
		})

		w.Header().Set("Content-Type", "application/json")

		if res.Phase == "ready" {
			logger.InfoWithMetadata("Notebook build completed", map[string]interface{}{
				"build_id":     buildId,
				"notebook_url": res.NotebookUrl.String,
				"remote_addr":  r.RemoteAddr,
			})

			newResponse := apiresponse.New("success", "Build completed redirect to the notebook url")
			dataResponse := newResponse.AddData(map[string]string{
				"token": res.Token.String,
				"url":   res.NotebookUrl.String,
				"phase": res.Phase,
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			return
		}

		// Handle other phases (building, failed, etc.)
		newResponse := apiresponse.New("pending", "Notebook Build In Progress")
		if res.Phase != "" {
			newResponse = apiresponse.New(res.Phase, "Notebook Status")
		}
		if res.Phase == "failed" {
			logger.WarnWithMetadata("Notebook build failed", map[string]interface{}{
				"build_id":    buildId,
				"message":     res.Message.String,
				"remote_addr": r.RemoteAddr,
			})
			newResponse = apiresponse.New(res.Phase, res.Message.String)
		}

		response, _ := newResponse.Marshal()
		w.Write(response)

		totalDuration := time.Since(start)
		logger.InfoWithMetadata("Notebook build status request completed", map[string]interface{}{
			"build_id":       buildId,
			"phase":          res.Phase,
			"total_duration": totalDuration.String(),
			"remote_addr":    r.RemoteAddr,
		})
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(getBuildStatus(app), middleware.LogRequest)
}
