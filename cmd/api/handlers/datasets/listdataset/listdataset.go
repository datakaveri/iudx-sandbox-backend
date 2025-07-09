package listdataset

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func listDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()
		ctx := r.Context()
		start := time.Now()

		logger.InfoWithContext(ctx, "Starting dataset list request")

		// Parse query parameters for enhanced functionality (optional)
		query := r.URL.Query()

		// Parse page parameter
		pageStr := query.Get("page")
		page := 1
		if pageStr != "" {
			if parsedPage, err := strconv.Atoi(pageStr); err != nil {
				logger.WarnWithMetadata("Invalid page parameter provided", map[string]interface{}{
					"page_param":   pageStr,
					"default_used": 1,
					"remote_addr":  r.RemoteAddr,
				})
			} else if parsedPage > 0 {
				page = parsedPage
			}
		}

		// Parse limit parameter
		limitStr := query.Get("limit")
		limit := 10 // default limit
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
				logger.WarnWithMetadata("Invalid limit parameter provided", map[string]interface{}{
					"limit_param":  limitStr,
					"default_used": 10,
					"remote_addr":  r.RemoteAddr,
				})
			} else if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			} else if parsedLimit > 100 {
				logger.WarnWithMetadata("Limit parameter exceeds maximum", map[string]interface{}{
					"requested_limit": parsedLimit,
					"max_limit":       100,
					"applied_limit":   100,
					"remote_addr":     r.RemoteAddr,
				})
				limit = 100
			}
		}

		// Parse search/filter parameters
		searchTerm := query.Get("search")
		domain := query.Get("domain")
		tag := query.Get("tag")

		logger.InfoWithMetadata("Dataset listing parameters parsed", map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"search_term": searchTerm,
			"domain":      domain,
			"tag":         tag,
			"remote_addr": r.RemoteAddr,
		})

		dataset := &models.Dataset{}

		// Use enhanced method if available, otherwise fallback to original
		var datasets interface{}
		var err error

		// Try enhanced method first (if it exists with GetAll)
		if searchTerm != "" || domain != "" || tag != "" || page > 1 || limit != 10 {
			logger.DebugWithMetadata("Using enhanced dataset query", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
			})
			datasets, err = dataset.GetAll(app, page, limit, searchTerm, domain, tag)
		} else {
			// Use original method for backward compatibility
			logger.DebugWithMetadata("Using original dataset query", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
			})
			datasets, err = dataset.ListDataset(app)
		}

		queryDuration := time.Since(start)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.InfoWithMetadata("No datasets found", map[string]interface{}{
					"remote_addr":    r.RemoteAddr,
					"query_duration": queryDuration.String(),
				})
				w.WriteHeader(http.StatusPreconditionFailed)
				return
			}

			logger.ErrorWithMetadata("Error in fetching datasets", map[string]interface{}{
				"remote_addr":    r.RemoteAddr,
				"error":          err.Error(),
				"query_duration": queryDuration.String(),
			}, err)

			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error in fetching datasets")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			return
		}

		// Log performance metrics
		if queryDuration > 5*time.Second {
			logger.WarnWithMetadata("Slow dataset query detected", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"remote_addr":    r.RemoteAddr,
			})
		} else if queryDuration > 1*time.Second {
			logger.InfoWithMetadata("Dataset query completed", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"remote_addr":    r.RemoteAddr,
			})
		} else {
			logger.DebugWithMetadata("Dataset query completed efficiently", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"remote_addr":    r.RemoteAddr,
			})
		}

		// Audit log for public data access
		logger.AuditLog(ctx, "dataset_list_accessed", "datasets", map[string]interface{}{
			"page":  page,
			"limit": limit,
			"filters": map[string]string{
				"search": searchTerm,
				"domain": domain,
				"tag":    tag,
			},
			"remote_addr": r.RemoteAddr,
		})

		totalDuration := time.Since(start)

		logger.InfoWithMetadata("Dataset list request completed successfully", map[string]interface{}{
			"remote_addr":    r.RemoteAddr,
			"total_duration": totalDuration.String(),
		})

		// IMPORTANT: Keep the exact same response format as original
		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "List of all datasets")

		dataResponse := newResponse.AddData(datasets)
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(listDataset(app), middleware.LogRequest)
}
