package router

import (
	"github.com/iudx-sandbox-backend/cmd/api/handlers/datasets/listdataset"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/buildnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/deletenotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/listnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/notebookbuildstatus"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/restartnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/stopnotebook"
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

	return mux
}
