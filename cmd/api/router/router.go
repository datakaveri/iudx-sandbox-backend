package router

import (
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/getdataset"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/listdataset"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/listdomains"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/listtags"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/onboarddataset"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/buildnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/deletenotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/listnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/notebookbuildstatus"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/restartnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/stopnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/referenceresources/listreferenceresource"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/referenceresources/onboardreferenceresource"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/resources/listresource"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/resources/onboardresource"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/health"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func Get(app *application.Application) *httprouter.Router {
	mux := httprouter.New()

	// Initialize health manager and checks
	healthManager := health.NewHealthManager("1.0.0") // TODO: get version from build
	healthManager.RegisterChecker("database", health.NewDatabaseChecker(app.DB.Client))
	healthManager.RegisterChecker("task_queue", health.NewTaskQueueChecker(app.TaskQueue))

	// Health check endpoints (public, no authentication required)
	mux.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		healthManager.HTTPHandler()(w, r)
	})
	mux.GET("/health/ready", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		healthManager.ReadinessHandler()(w, r)
	})
	mux.GET("/health/live", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		healthManager.LivenessHandler()(w, r)
	})

	// Notebook endpoints - require authentication for all operations
	mux.GET("/api/notebooks", middleware.Chain(
		listnotebook.Do(app),
		middleware.AuthorizeRequest,
	))
	mux.GET("/api/notebooks/build-status", middleware.Chain(
		notebookbuildstatus.Do(app),
	))
	mux.POST("/api/notebooks", middleware.Chain(
		buildnotebook.Do(app),
		middleware.AuthorizeRequest,
	))
	mux.DELETE("/api/notebooks", middleware.Chain(
		deletenotebook.Do(app),
		middleware.AuthorizeRequest,
	))
	mux.GET("/api/notebooks/stop", middleware.Chain(
		stopnotebook.Do(app),
		middleware.AuthorizeRequest,
	))
	mux.GET("/api/notebooks/start", middleware.Chain(
		restartnotebook.Do(app),
		middleware.AuthorizeRequest,
	))

	// Dataset endpoints - read operations for consumers, onboard operations use static API key
	mux.GET("/api/datasets", middleware.Chain(
		listdataset.Do(app),
	))
	mux.POST("/api/dataset", middleware.Chain(
		onboarddataset.Do(app),
		middleware.AuthorizeStaticAPIKey, // Service-to-service authentication
	))
	mux.GET("/api/dataset/:id", middleware.Chain(
		getdataset.Do(app),
	))

	// Resource endpoints - read operations for consumers, onboard operations use static API key
	mux.GET("/api/resources/:id", middleware.Chain(
		listresource.Do(app),
	))
	mux.POST("/api/resource", middleware.Chain(
		onboardresource.Do(app),
		middleware.AuthorizeStaticAPIKey, // Service-to-service authentication
	))

	// Reference resource endpoints - read operations for consumers, onboard operations use static API key
	mux.GET("/api/referenceresources/:id", middleware.Chain(
		listreferenceresource.Do(app),
	))
	mux.POST("/api/referenceresource", middleware.Chain(
		onboardreferenceresource.Do(app),
		middleware.AuthorizeStaticAPIKey, // Service-to-service authentication
	))

	// Metadata endpoints - read-only, require basic authentication
	mux.GET("/api/tags", middleware.Chain(
		listtags.Do(app),
	))
	mux.GET("/api/domains", middleware.Chain(
		listdomains.Do(app),
	))

	return mux
}
