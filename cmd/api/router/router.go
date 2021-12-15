package router

import (
	"io"
	"net/http"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/julienschmidt/httprouter"
)

func Get(app *application.Application) *httprouter.Router {
	mux := httprouter.New()

	mux.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "Hello from Sandbox!")
	})

	return mux
}
