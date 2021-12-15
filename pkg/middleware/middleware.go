package middleware

import (
	"github.com/julienschmidt/httprouter"
)

type Middleware func(httprouter.Handle) httprouter.Handle

func Chain(f httprouter.Handle, m ...Middleware) httprouter.Handle {
	if len(m) == 0 {
		return f
	}

	return m[0](Chain(f, m[1:]...))
}
