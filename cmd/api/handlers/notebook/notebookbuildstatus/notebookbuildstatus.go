package notebookbuildstatus

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func getBuildStatus(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		queryValues := r.URL.Query()

		notebook := &models.Notebook{}
		res, err := notebook.GetBuildStatus(app, queryValues.Get("buildId"))

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNoContent)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error in fetching Notebook %v\n", err)
			return
		}

		if res.Phase == "ready" {
			w.Header().Set("Content-Type", "application/json")
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

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("pending", "Notebook Build In Progress")
		if res.Phase != "" {
			newResponse = apiresponse.New(res.Phase, "Notebook Status")
		}
		if res.Phase == "failed" {
			newResponse = apiresponse.New(res.Phase, res.Message.String)
		}
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(getBuildStatus(app), middleware.LogRequest)
}
