package listtags

import (
	"net/http"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/julienschmidt/httprouter"
)

func listTags(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer r.Body.Close()

		// tags := &models.Tag{}

		// tags, err := tags.ListTags(app)
	}
}
