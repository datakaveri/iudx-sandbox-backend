package main

import (
	"context"
	"time"

	"github.com/iudx-sandbox-backend/cmd/api/router"
	"github.com/iudx-sandbox-backend/cmd/tasks"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/exithandler"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/iudx-sandbox-backend/pkg/server"
	"github.com/joho/godotenv"
)

func main() {
	startTime := time.Now()

	logger.InfoWithMetadata("Starting IUDX Sandbox Backend application", map[string]interface{}{
		"start_time": startTime.Format(time.RFC3339),
		"go_version": "1.24",
	})

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.WarnWithMetadata("Failed to load .env file, using system environment variables", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.DebugWithMetadata("Environment file loaded successfully", map[string]interface{}{
			"file": ".env",
		})
	}

	// Initialize JWT configuration
	logger.DebugWithMetadata("Initializing JWT authentication configuration", map[string]interface{}{
		"component": "authutility",
	})

	if err := authutility.InitJWTConfig(); err != nil {
		logger.FatalWithError("Failed to initialize JWT configuration", err)
	}

	logger.InfoWithMetadata("JWT configuration initialized successfully", map[string]interface{}{
		"component": "authutility",
		"status":    "ready",
	})

	// Initialize application
	logger.DebugWithMetadata("Initializing application components", map[string]interface{}{
		"components": []string{"database", "config", "task_queue"},
	})

	app, err := application.Get()
	if err != nil {
		logger.FatalWithError("Failed to initialize application", err)
	}

	logger.InfoWithMetadata("Application components initialized successfully", map[string]interface{}{
		"components_ready":    []string{"database", "config", "task_queue"},
		"initialization_time": time.Since(startTime).String(),
	})

	// Validate configuration
	logger.DebugWithMetadata("Validating application configuration", map[string]interface{}{
		"environment": app.Cfg.GetEnvironment(),
	})

	if err := app.Cfg.ValidateConfig(); err != nil {
		logger.FatalWithError("Configuration validation failed", err)
	}

	logger.InfoWithMetadata("Configuration validation passed", map[string]interface{}{
		"environment": app.Cfg.GetEnvironment(),
		"api_port":    app.Cfg.GetAPIPort(),
		"status":      "validated",
	})

	// Start the in-memory task queue
	logger.InfoWithMetadata("Starting task queue system", map[string]interface{}{
		"queue_type": "in-memory",
	})

	ctx := context.Background()
	if err := app.TaskQueue.Start(ctx); err != nil {
		logger.FatalWithError("Failed to start task queue", err)
	}

	logger.InfoWithMetadata("Task queue started successfully", map[string]interface{}{
		"queue_type": "in-memory",
		"status":     "running",
	})

	// Configure server
	logger.DebugWithMetadata("Configuring HTTP server", map[string]interface{}{
		"port":        app.Cfg.GetAPIPort(),
		"environment": app.Cfg.GetEnvironment(),
	})

	srv := server.
		Get().
		WithAddr(app.Cfg.GetAPIPort()).
		WithRouter(app.Cfg.GetEnvironment(), router.Get(app)).
		WithErrLogger(logger.Error)

	// Start HTTP server in background
	go func() {
		serverStartTime := time.Now()

		logger.InfoWithMetadata("Starting HTTP server", map[string]interface{}{
			"port":         app.Cfg.GetAPIPort(),
			"environment":  app.Cfg.GetEnvironment(),
			"server_start": serverStartTime.Format(time.RFC3339),
		})

		if err := srv.Start(); err != nil {
			logger.FatalWithError("HTTP server failed to start", err)
		}
	}()

	// Start background task workers
	logger.InfoWithMetadata("Starting background task workers", map[string]interface{}{
		"component": "task_workers",
	})

	tasks.StartTask(app)

	logger.InfoWithMetadata("Background task workers started", map[string]interface{}{
		"component": "task_workers",
		"status":    "running",
	})

	// Log startup completion
	totalStartupTime := time.Since(startTime)
	logger.InfoWithMetadata("Application startup completed successfully", map[string]interface{}{
		"total_startup_time": totalStartupTime.String(),
		"components_running": []string{
			"http_server",
			"task_queue",
			"task_workers",
			"database",
		},
		"environment": app.Cfg.GetEnvironment(),
		"port":        app.Cfg.GetAPIPort(),
		"ready":       true,
	})

	// Log startup performance
	if totalStartupTime > 30*time.Second {
		logger.WarnWithMetadata("Application startup was very slow", map[string]interface{}{
			"startup_time": totalStartupTime.String(),
			"threshold":    "30s",
		})
	} else if totalStartupTime > 10*time.Second {
		logger.InfoWithMetadata("Application startup completed", map[string]interface{}{
			"startup_time": totalStartupTime.String(),
			"performance":  "acceptable",
		})
	} else {
		logger.DebugWithMetadata("Application startup was fast", map[string]interface{}{
			"startup_time": totalStartupTime.String(),
			"performance":  "excellent",
		})
	}

	// Initialize graceful shutdown
	logger.DebugWithMetadata("Initializing graceful shutdown handler", map[string]interface{}{
		"signals": []string{"SIGINT", "SIGTERM"},
	})

	exithandler.Init(func() {
		shutdownStart := time.Now()

		logger.InfoWithMetadata("Initiating graceful shutdown", map[string]interface{}{
			"shutdown_start": shutdownStart.Format(time.RFC3339),
			"reason":         "signal_received",
		})

		// Stop task queue gracefully
		logger.DebugWithMetadata("Stopping task queue", map[string]interface{}{
			"component": "task_queue",
		})

		if err := app.TaskQueue.Stop(); err != nil {
			logger.ErrorWithMetadata("Error stopping task queue during shutdown", map[string]interface{}{
				"component": "task_queue",
			}, err)
		} else {
			logger.InfoWithMetadata("Task queue stopped successfully", map[string]interface{}{
				"component": "task_queue",
				"status":    "stopped",
			})
		}

		// Stop HTTP server
		logger.DebugWithMetadata("Stopping HTTP server", map[string]interface{}{
			"component": "http_server",
		})

		if err := srv.Close(); err != nil {
			logger.ErrorWithMetadata("Error stopping HTTP server during shutdown", map[string]interface{}{
				"component": "http_server",
			}, err)
		} else {
			logger.InfoWithMetadata("HTTP server stopped successfully", map[string]interface{}{
				"component": "http_server",
				"status":    "stopped",
			})
		}

		// Close database connections
		logger.DebugWithMetadata("Closing database connections", map[string]interface{}{
			"component": "database",
		})

		if err := app.DB.Close(); err != nil {
			logger.ErrorWithMetadata("Error closing database connections during shutdown", map[string]interface{}{
				"component": "database",
			}, err)
		} else {
			logger.InfoWithMetadata("Database connections closed successfully", map[string]interface{}{
				"component": "database",
				"status":    "closed",
			})
		}

		shutdownDuration := time.Since(shutdownStart)
		logger.InfoWithMetadata("Graceful shutdown completed", map[string]interface{}{
			"shutdown_duration": shutdownDuration.String(),
			"total_uptime":      time.Since(startTime).String(),
			"status":            "shutdown_complete",
		})
	})
}
