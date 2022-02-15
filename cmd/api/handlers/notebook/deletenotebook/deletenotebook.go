package deletenotebook

import (
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

func deleteNotebook(app *application.Application) httprouter.Handle {
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

		// Remove spawner id from notebook table
		if err := notebook.RemoveSpawnerId(app); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error deleting notebook %v\n", err)
			return
		}

		jupyterClient, err := jupyterutility.Get()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure", err)
			return
		}

		spawnerName, err := notebook.GetSpawnerName(app, user.UserId, notebookId)

		_, err = jupyterClient.DeleteServer(app, tokenUser.UserName, spawnerName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Jupyter client failure", err)
			return
		}

		// Delete notebook entry
		if err := notebook.Delete(app, notebookId); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error deleting notebook %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Notebook Deleted")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(deleteNotebook(app), middleware.LogRequest)
}
