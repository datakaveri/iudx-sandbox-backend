package listresource

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

func listResource(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		unique_id := p.ByName("id")

		defer r.Body.Close()

		resource := &models.Resource{}

		resources, err := resource.ListResource(app, unique_id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error in fetching resources")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			logger.Error.Printf("Error in fetching Datasets %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "List of all resources")

		dataResponse := newResponse.AddData(resources)
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(listResource(app), middleware.LogRequest)
}
