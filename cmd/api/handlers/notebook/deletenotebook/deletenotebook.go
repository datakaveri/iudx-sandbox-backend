package deletenotebook

import (
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func deleteNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		queryValues := r.URL.Query()

		notebook := &models.Notebook{}
		notebookId := queryValues.Get("notebookId")
		// TODO delete stopped container or stop and then delete

		if err := notebook.Delete(app, notebookId); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error deleting notebook %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook Deleted")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(deleteNotebook(app), middleware.LogRequest)
}
