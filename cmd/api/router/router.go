package router

import (
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
	"github.com/julienschmidt/httprouter"
)

func Get(app *application.Application) *httprouter.Router {
	mux := httprouter.New()

	mux.GET("/api/notebooks", listnotebook.Do(app))
	mux.GET("/api/notebooks/build-status", notebookbuildstatus.Do(app))
	mux.POST("/api/notebooks", buildnotebook.Do(app))
	mux.DELETE("/api/notebooks", deletenotebook.Do(app))
	mux.GET("/api/notebooks/stop", stopnotebook.Do(app))
	mux.GET("/api/notebooks/start", restartnotebook.Do(app))

	mux.GET("/api/datasets", listdataset.Do(app))
	mux.POST("/api/dataset", onboarddataset.Do(app))
	mux.GET("/api/dataset/:id", getdataset.Do(app))

	mux.GET("/api/resources/:id", listresource.Do(app))
	mux.POST("/api/resource", onboardresource.Do(app))

	mux.GET("/api/referenceresources/:id", listreferenceresource.Do(app))
	mux.POST("/api/referenceresource", onboardreferenceresource.Do(app))

	mux.GET("/api/tags", listtags.Do(app))
	mux.GET("/api/domains", listdomains.Do(app))

	return mux
}
