package onboarddataset

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

func onboardDataset(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		// Authenticate user (optional, depending on your requirements)
		// _, err := authutility.ExtractTokenMetadata(r)
		// if err != nil {
		// 	logger.Error.Printf("Error in deleting dataset, Unauthorized %v\n", err)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		defer r.Body.Close()

		dataset := &models.Dataset{}
		json.NewDecoder(r.Body).Decode(dataset)

		if err := dataset.OnboardDataset(app); err != nil {
			logger.Error.Printf("Error in onboarding dataset %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Dataset was not inserted")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			return
		}

		w.Header().Set("Content-Type", "application-json")
		newResponse := apiresponse.New("success", "Dataset onboarded successfully")
		dataResponse := newResponse.AddData(map[string]string{
			"Success": "True",
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(onboardDataset(app), middleware.LogRequest)
}
