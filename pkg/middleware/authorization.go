package middleware

import (
	"net/http"

	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/julienschmidt/httprouter"
)

func AuthorizeRequest(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := authutility.TokenValid(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			next(w, r, p)
		}
	}
}
