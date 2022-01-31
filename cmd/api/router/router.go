package router

import (
	"github.com/iudx-sandbox-backend/cmd/api/handlers/gallery/createrepo"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/gallery/deleterepo"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/gallery/getrepo"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/gallery/listrepo"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/buildnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/deletenotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/listnotebook"
	"github.com/iudx-sandbox-backend/cmd/api/handlers/notebook/notebookbuildstatus"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/julienschmidt/httprouter"
)

func Get(app *application.Application) *httprouter.Router {
	mux := httprouter.New()

	mux.GET("/api/gallery", listrepo.Do(app))
	mux.GET("/api/gallery/:id", getrepo.Do(app))
	mux.POST("/api/gallery", createrepo.Do(app))
	mux.DELETE("/api/gallery/:id", deleterepo.Do(app))

	mux.GET("/api/notebooks", listnotebook.Do(app))
	mux.GET("/api/notebooks/build-status", notebookbuildstatus.Do(app))
	mux.POST("/api/notebooks", buildnotebook.Do(app))
	mux.DELETE("/api/notebooks", deletenotebook.Do(app))

	return mux
}
