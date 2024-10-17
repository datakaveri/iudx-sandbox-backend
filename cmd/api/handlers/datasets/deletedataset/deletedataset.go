package deletedataset

import (
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func deleteDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		// Extract dataset ID from URL parameters
		datasetID := p.ByName("id")

		// Authenticate user (optional, depending on your requirements)
		// _, err := authutility.ExtractTokenMetadata(r)
		// if err != nil {
		// 	logger.Error.Printf("Error in deleting dataset, Unauthorized %v\n", err)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// Initialize dataset model
		dataset := &models.Dataset{}

		// Delete the dataset
		if err := dataset.Delete(app, datasetID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error deleting dataset")
			badResponse := newResponse.AddReason(err.Error())
			response, _ := badResponse.Marshal()
			w.Write(response)
			logger.Error.Printf("Error deleting dataset: %v\n", err)
			return
		}

		// Respond with success message
		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Dataset deleted successfully")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(deleteDataset(app), middleware.LogRequest)
}
