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
		defer r.Body.Close()

		dataset := &models.Dataset{}
		json.NewDecoder(r.Body).Decode(dataset)

		if err := dataset.Onboard(app); err != nil {
			logger.Error.Printf("Error in creating dataset %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application-json")
		newResponse := apiresponse.New("success", "Dataset inserted successfully")
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
