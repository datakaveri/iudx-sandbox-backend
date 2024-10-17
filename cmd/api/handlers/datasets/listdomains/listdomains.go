package listdomains

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

func listDomains(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Authenticate user (optional, depending on your requirements)
		// _, err := authutility.ExtractTokenMetadata(r)
		// if err != nil {
		// 	logger.Error.Printf("Error in deleting dataset, Unauthorized %v\n", err)
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		defer r.Body.Close()

		domain := &models.Tag{}

		domains, err := domain.ListDomains(app)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
			}
			w.WriteHeader(http.StatusInternalServerError)
			newResponse := apiresponse.New("failed", "Error in fetching domains")
			dataResponse := newResponse.AddData(map[string]string{
				"Error": err.Error(),
			})
			response, _ := dataResponse.Marshal()
			w.Write(response)
			logger.Error.Printf("Error in fetching domains %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "List of all domains")

		dataResponse := newResponse.AddData(domains)
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(listDomains(app), middleware.LogRequest)
}
