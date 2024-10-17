package updatedataset

import (
	"encoding/json"
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func updateDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		// Extract dataset ID from URL parameters
		datasetID := p.ByName("id")

		// Authenticate user (optional, depending on your requirements)
		// _, err := authutility.ExtractTokenMetadata(r)
		// if err != nil {
		// 	logger.Error.Printf("Error in updating dataset, Unauthorized %v\n", err)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// Parse the request body
		var updateData map[string]interface{}

		err := json.NewDecoder(r.Body).Decode(&updateData)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			newResponse := apiresponse.New("failed", "Invalid request body")
			badResponse := newResponse.AddReason(err.Error())
			response, _ := badResponse.Marshal()
			w.Write(response)
			return
		}

		// Update the dataset
		if err := models.UpdateDataset(app, datasetID, updateData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error updating dataset")
			badResponse := newResponse.AddReason(err.Error())
			response, _ := badResponse.Marshal()
			w.Write(response)
			logger.Error.Printf("Error updating dataset: %v\n", err)
			return
		}

		// Respond with success message
		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Dataset updated successfully")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(updateDataset(app), middleware.LogRequest)
}
