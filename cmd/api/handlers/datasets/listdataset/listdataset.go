package listdataset

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

func listDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		dataset := &models.Dataset{}

		datasets, err := dataset.Get(app)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error in fetching Notebooks %v\n", err)
			return
		}

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
