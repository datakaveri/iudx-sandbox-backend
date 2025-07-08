package tasks

import (
	"time"

	"github.com/iudx-sandbox-backend/cmd/tasks/handlers/restartnotebook"
	"github.com/iudx-sandbox-backend/cmd/tasks/handlers/spawnernotebooksync"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

func StartTask(app *application.Application) {
	// Register task handlers
	spawnernotebooksync.RegisterTask(app)
	restartnotebook.RegisterTask(app)

	// Start monitoring goroutine for task queue statistics
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for range ticker.C {
			if app.TaskQueue.IsStarted() {
				stats := app.TaskQueue.GetStats()
				logger.Info.Printf("Task Queue Stats - Processed: %d, Failed: %d, Active: %d, Queued: %d",
					stats.Processed, stats.Failed, stats.Active, stats.Queued)

				// Check queue health
				if err := app.TaskQueue.Health(); err != nil {
					logger.Error.Printf("Task queue health check failed: %v", err)
				}
			}
		}
	}()

	logger.Info.Printf("Task monitoring started with in-memory task queue")
}
