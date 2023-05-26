package tasks

import (
	"context"
	"time"

	// "github.com/iudx-sandbox-backend/cmd/tasks/handlers/spawnernotebooksync"
	"github.com/iudx-sandbox-backend/pkg/application"
	// "github.com/iudx-sandbox-backend/pkg/exithandler"
	"github.com/iudx-sandbox-backend/pkg/logger"
	// "github.com/joho/godotenv"
)

func StartTask(app *application.Application) {
	// if err := godotenv.Load(); err != nil {
	// 	logger.Info.Println("Failed to load env file")
	// }

	// app, err := application.Get()

	// if err != nil {
	// 	logger.Error.Printf("Error creating worker application %v\n", err)
	// }

	// spawnernotebooksync.RegisterTask()

	go func() {
		ctx := context.Background()

		if err := app.TaskQueue.Queue.Consumer().Start(ctx); err != nil {
			logger.Error.Println(err)
		}

		logger.Info.Printf("Starting task consumers")
	}()

	go func() {
		for range time.Tick(2 * time.Second) {
			stats := app.TaskQueue.Queue.Consumer().Stats()

			logger.Info.Printf("Number of process workers %d\n", stats.NumWorker)
			logger.Info.Printf("Number of tasks processed %d\n", stats.Processed)
		}
	}()

	// exithandler.Init(func() {
	// 	if err := app.TaskQueue.Queue.Consumer().Stop(); err != nil {
	// 		logger.Error.Println(err.Error())
	// 	}

	// 	if err := app.DB.Close(); err != nil {
	// 		logger.Error.Println(err.Error())
	// 	}
	// })
}
