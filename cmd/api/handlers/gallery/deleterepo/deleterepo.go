package deleterepo

import (
	"net/http"
	"strings"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/apiresponse"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/middleware"
	"github.com/julienschmidt/httprouter"
)

func deleteRepo(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		repoName := strings.TrimPrefix(r.URL.Path, "/api/gallery/")

		repo := &models.Repo{}

		if err := repo.Delete(app, repoName); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Printf("Error deleting Gallery Repo %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		newResponse := apiresponse.New("success", "Deleted Gallery Repo")
		response, _ := newResponse.Marshal()
		w.Write(response)
	}
}

func Do(app *application.Application) httprouter.Handle {
	return middleware.Chain(deleteRepo(app), middleware.LogRequest)
}
