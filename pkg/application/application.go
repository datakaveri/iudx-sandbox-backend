package application

import (
	"github.com/iudx-sandbox-backend/pkg/config"
)

type Application struct {
	Cfg *config.Config
}

func Get() (*Application, error) {
	cfg := config.Get()

	return &Application{
		Cfg: cfg,
	}, nil
}
