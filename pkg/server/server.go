package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

type Server struct {
	srv *http.Server
}

func Get() *Server {
	return &Server{
		srv: &http.Server{},
	}
}

func (s *Server) WithAddr(addr string) *Server {
	s.srv.Addr = addr
	return s
}

func (s *Server) WithErrLogger(l *log.Logger) *Server {
	s.srv.ErrorLog = l
	return s
}

func (s *Server) WithRouter(router *httprouter.Router) *Server {
	s.srv.Handler = cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:4500", "*"},
		AllowedMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization", "Content-Type", "BuildToken", "token"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		// Debug:              true,
	}).Handler(router)
	return s
}

func (s *Server) Start() error {
	if len(s.srv.Addr) == 0 {
		return errors.New("Server missing address")
	}

	if s.srv.Handler == nil {
		return errors.New("Server missing handler")
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Close() error {
	return s.srv.Close()
}
