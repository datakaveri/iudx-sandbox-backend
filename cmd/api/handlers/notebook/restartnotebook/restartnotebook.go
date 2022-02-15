package restartnotebook

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/jupyterutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func restartNotebook(app *application.Application) httprouter.Handle {
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

		queryValues := r.URL.Query()
		notebook := &models.Notebook{}
		notebookId := queryValues.Get("notebookId")
		notebook.NotebookId = notebookId

		jupyterClient, err := jupyterutility.Get()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure")
			return
		}

		spawnerName, err := notebook.GetSpawnerName(app, user.UserId, notebookId)

		_, err = jupyterClient.RestartServer(app, tokenUser.UserName, spawnerName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure", err)
			return
		}

		notebook.Phase = "ready"
		err = notebook.UpdateNotebookStatus(app)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusPreconditionFailed)
				logger.Info.Println("No records found")
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error in restarting notebook %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook starting")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(restartNotebook(app), middleware.LogRequest, middleware.AuthorizeRequest)
}
