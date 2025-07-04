package listdataset

import (
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
		ctx := r.Context()
		start := time.Now()

		logger.InfoWithContext(ctx, "Starting dataset list request")

		// Parse query parameters
		query := r.URL.Query()

		// Parse page parameter
		pageStr := query.Get("page")
		page := 1
		if pageStr != "" {
			if parsedPage, err := strconv.Atoi(pageStr); err != nil {
				logger.WarnWithMetadata("Invalid page parameter provided", map[string]interface{}{
					"page_param":   pageStr,
					"default_used": 1,
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
				})
			} else if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			} else if parsedLimit > 100 {
				logger.WarnWithMetadata("Limit parameter exceeds maximum", map[string]interface{}{
					"requested_limit": parsedLimit,
					"max_limit":       100,
					"applied_limit":   100,
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

		// Create dataset model and perform query
		datasetModel := &models.Dataset{}

		logger.DebugWithMetadata("Executing dataset query", map[string]interface{}{
			"query_params": map[string]interface{}{
				"page":   page,
				"limit":  limit,
				"search": searchTerm,
				"domain": domain,
				"tag":    tag,
			},
			"remote_addr": r.RemoteAddr,
		})

		datasets, err := datasetModel.GetAll(app, page, limit, searchTerm, domain, tag)
		queryDuration := time.Since(start)

		if err != nil {
			logger.ErrorWithMetadata("Database query failed for dataset listing", map[string]interface{}{
				"page":           page,
				"limit":          limit,
				"search_term":    searchTerm,
				"domain":         domain,
				"tag":            tag,
				"query_duration": queryDuration.String(),
				"remote_addr":    r.RemoteAddr,
			}, err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Log performance metrics
		if queryDuration > 5*time.Second {
			logger.WarnWithMetadata("Slow dataset query detected", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"result_count":   len(datasets),
				"page":           page,
				"limit":          limit,
				"remote_addr":    r.RemoteAddr,
			})
		} else if queryDuration > 1*time.Second {
			logger.InfoWithMetadata("Dataset query completed", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"result_count":   len(datasets),
				"remote_addr":    r.RemoteAddr,
			})
		} else {
			logger.DebugWithMetadata("Dataset query completed efficiently", map[string]interface{}{
				"query_duration": queryDuration.String(),
				"result_count":   len(datasets),
				"remote_addr":    r.RemoteAddr,
			})
		}

		// Audit log for public data access
		logger.AuditLog(ctx, "dataset_list_accessed", "datasets", map[string]interface{}{
			"result_count": len(datasets),
			"page":         page,
			"limit":        limit,
			"filters": map[string]string{
				"search": searchTerm,
				"domain": domain,
				"tag":    tag,
			},
			"remote_addr": r.RemoteAddr,
		})

		// Prepare response
		w.Header().Set("Content-Type", "application/json")

		newResponse := apiresponse.New("success", "Datasets retrieved successfully")
		dataResponse := newResponse.AddData(map[string]interface{}{
			"datasets": datasets,
			"pagination": map[string]interface{}{
				"page":         page,
				"limit":        limit,
				"result_count": len(datasets),
			},
		})

		response, err := dataResponse.Marshal()
		if err != nil {
			logger.ErrorWithMetadata("Failed to marshal dataset response", map[string]interface{}{
				"result_count": len(datasets),
				"remote_addr":  r.RemoteAddr,
			}, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		totalDuration := time.Since(start)

		logger.InfoWithMetadata("Dataset list request completed successfully", map[string]interface{}{
			"total_duration": totalDuration.String(),
			"result_count":   len(datasets),
			"response_size":  len(response),
			"remote_addr":    r.RemoteAddr,
		})

		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(listDataset(app), middleware.LogRequest)
}
