package listnotebook

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func listNotebook(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()
		tokenUser, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.Error.Printf("Error in fetching notebook, Unauthorized %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userModel := &models.User{}
		user, err := userModel.Get(app, tokenUser.UserName)

		notebook := &models.Notebook{}
		notebooks, err := notebook.List(app, user.UserId)

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
		newResponse := apiresponse.New("success", "List of all notebooks")
		dataResponse := newResponse.AddData(notebooks)
		response, _ := dataResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(listNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
