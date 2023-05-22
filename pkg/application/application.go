package application

import (
	"github.com/iudx-sandbox-backend/pkg/config"
	"github.com/iudx-sandbox-backend/pkg/db"
	"github.com/iudx-sandbox-backend/pkg/taskqueue"
)

type Application struct {
	DB        *db.DB
	Cfg       *config.Config
	TaskQueue *taskqueue.TaskQueue
}

func Get() (*Application, error) {
	cfg := config.Get()

	db, err := db.Get(cfg.GetDBConnStr())
	if err != nil {
		return nil, err
	}

	tq := taskqueue.InitializeTaskQueue()

	return &Application{
		DB:        db,
		Cfg:       cfg,
		TaskQueue: tq,
	}, nil
}
