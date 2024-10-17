package getdataset

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

func getDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		unique_id := p.ByName("id")

		// Authenticate user (optional, depending on your requirements)
		// _, err := authutility.ExtractTokenMetadata(r)
		// if err != nil {
		// 	logger.Error.Printf("Error in deleting dataset, Unauthorized %v\n", err)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		defer r.Body.Close()

		dataset := &models.Dataset{}

		datasets, err := dataset.GetDataset(app, unique_id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error in fetching datasets")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			logger.Error.Printf("Error in fetching Datasets %v\n", err)
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
	return middleware.Chain(getDataset(app), middleware.LogRequest)
}
