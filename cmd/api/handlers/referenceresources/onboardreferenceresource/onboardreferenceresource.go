package onboardreferenceresource

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

func onboardReferenceResource(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		referenceResource := &models.ReferenceResource{}
		json.NewDecoder(r.Body).Decode(referenceResource)

		if err := referenceResource.OnboardReferenceResource(app); err != nil {
			logger.Error.Printf("Error in onboarding resource %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Reference Resource was not onboarded")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			return
		}

		w.Header().Set("Content-Type", "application-json")
		newResponse := apiresponse.New("success", "Reference Resource onboarded successfully")
		dataResponse := newResponse.AddData(map[string]string{
			"Success": "True",
		})
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(onboardReferenceResource(app), middleware.LogRequest)
}
