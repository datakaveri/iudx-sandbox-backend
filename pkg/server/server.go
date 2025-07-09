package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

type Server struct {
	srv *http.Server
}

func Get() *Server {
	return &Server{
		srv: &http.Server{
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
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

func (s *Server) WithRouter(environment string, router *httprouter.Router) *Server {
	// Get allowed origins from environment variable
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	if allowedOriginsEnv != "" {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
		// Trim whitespace from each origin
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// Default secure origins based on environment
		switch environment {
		case "dev", "development":
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:4500",
				"http://localhost:8080",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:4500",
				"http://127.0.0.1:8080",
			}
		case "staging":
			allowedOrigins = []string{
				"https://staging.iudx.org.in",
				"https://sandbox-staging.iudx.org.in",
			}
		case "prod", "production":
			allowedOrigins = []string{
				"https://iudx.org.in",
				"https://sandbox.iudx.org.in",
			}
		default:
			// Most restrictive - only same origin
			allowedOrigins = []string{}
		}
	}

	// Configure CORS with security best practices
	corsOptions := cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
			"Accept",
			"Origin",
			"User-Agent",
			"Cache-Control",
			"Keep-Alive",
			"X-Requested-With",
			"If-Modified-Since",
			"X-CSRF-Token",
			// Legacy header for backward compatibility
			"token",
			"BuildToken",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials:   true,
		MaxAge:             int(12 * time.Hour / time.Second), // 12 hours
		OptionsPassthrough: false,                             // Handle preflight requests properly
		Debug:              environment == "dev" || environment == "development",
	}

	// Apply CORS middleware
	s.srv.Handler = cors.New(corsOptions).Handler(router)

	return s
}

func (s *Server) Start() error {
	if len(s.srv.Addr) == 0 {
		return errors.New("server missing address")
	}

	if s.srv.Handler == nil {
		return errors.New("server missing handler")
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Close() error {
	return s.srv.Close()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
