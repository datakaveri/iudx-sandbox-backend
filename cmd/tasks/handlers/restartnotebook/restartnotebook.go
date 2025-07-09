package restartnotebook

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/r3labs/sse/v2"
)

type RestartNotebookTaskPayload struct {
	ProgressUrl string `json:"progress_url"`
	BuildId     string `json:"build_id"`
	UserId      int    `json:"user_id"`
}

type RestartProgressResponse struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

func customHeader(app *application.Application) func(c *sse.Client) {
	return func(c *sse.Client) {
		c.Headers = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", app.Cfg.GetJupyterHubApiToken()),
		}
	}
}

func handleRestartSse(app *application.Application, progressUrl, buildId string, userId int) error {
	start := time.Now()

	logger.InfoWithMetadata("Starting SSE connection for notebook restart", map[string]interface{}{
		"build_id":     buildId,
		"user_id":      userId,
		"progress_url": progressUrl,
	})

	if progressUrl == "" {
		err := fmt.Errorf("progress URL is empty")
		logger.ErrorWithMetadata("SSE connection failed - missing progress URL", map[string]interface{}{
			"build_id": buildId,
			"user_id":  userId,
		}, err)
		return err
	}

	sseClient := sse.NewClient(progressUrl, customHeader(app))

	// Disable TLS certificate verification to handle self-signed or invalid certificates
	sseClient.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	logger.InfoWithMetadata("SSE client created for restart with TLS verification disabled", map[string]interface{}{
		"build_id":     buildId,
		"progress_url": progressUrl,
	})

	eventCount := 0
	lastEventTime := time.Now()
	done := make(chan struct{})
	var sseErr error

	logger.InfoWithMetadata("Attempting to subscribe to SSE events for restart", map[string]interface{}{
		"build_id":     buildId,
		"progress_url": progressUrl,
	})

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		eventCount++
		eventTime := time.Now()
		timeSinceLastEvent := eventTime.Sub(lastEventTime)
		lastEventTime = eventTime

		logger.InfoWithMetadata("SSE restart event received", map[string]interface{}{
			"build_id":              buildId,
			"event_count":           eventCount,
			"time_since_last_event": timeSinceLastEvent.String(),
			"event_data_size":       len(msg.Data),
		})

		logger.DebugWithMetadata("Raw SSE restart event data", map[string]interface{}{
			"build_id": buildId,
			"raw_data": string(msg.Data),
		})

		response := &RestartProgressResponse{}
		if err := json.Unmarshal(msg.Data, response); err != nil {
			logger.ErrorWithMetadata("Failed to unmarshal SSE restart event data", map[string]interface{}{
				"build_id":    buildId,
				"event_count": eventCount,
				"raw_data":    string(msg.Data),
			}, err)
			return
		}

		logger.InfoWithMetadata("Notebook restart status update", map[string]interface{}{
			"build_id":    buildId,
			"ready":       response.Ready,
			"message":     response.Message,
			"url":         response.Url,
			"event_count": eventCount,
		})

		notebook := &models.Notebook{}
		notebook.BuildId = buildId
		notebook.Message = response.Message

		if response.Ready {
			logger.InfoWithMetadata("Notebook restart completed - processing spawner ID", map[string]interface{}{
				"build_id": buildId,
				"url":      response.Url,
			})

			// Find and update spawner ID
			spawner := &models.Spawner{}
			res, err := spawner.GetSpawnerIdBasedOnBaseUrl(app, response.Url, userId)
			if err != nil {
				logger.ErrorWithMetadata("Error finding spawner id during restart", map[string]interface{}{
					"build_id": buildId,
					"url":      response.Url,
					"user_id":  userId,
				}, err)
			} else {
				notebook.SpawnerId = res.Id
				if err := notebook.UpdateNotebookSpawnerId(app); err != nil {
					logger.ErrorWithMetadata("Error failed to update spawner id during restart", map[string]interface{}{
						"build_id":   buildId,
						"spawner_id": notebook.SpawnerId,
					}, err)
				}
			}

			notebook.Phase = "ready"
			// Construct full notebook URL from JupyterHub base URL and relative path
			notebookUrl := ""
			if response.Url != "" {
				// Get JupyterHub API base URL (e.g., "https://hub.playground.iudx.org.in/hub/api")
				jupyterHubApiUrl := app.Cfg.GetJupyterHubApi()

				// Remove /hub/api suffix to get the base URL
				baseUrl := jupyterHubApiUrl
				if len(baseUrl) >= 8 && baseUrl[len(baseUrl)-8:] == "/hub/api" {
					baseUrl = baseUrl[:len(baseUrl)-8]
				}

				// Ensure response.Url starts with /
				relativeUrl := response.Url
				if !strings.HasPrefix(relativeUrl, "/") {
					relativeUrl = "/" + relativeUrl
				}

				// Remove /lab suffix if present to avoid duplication
				relativeUrl = strings.TrimSuffix(relativeUrl, "/lab")

				// Construct full URL: baseUrl + relativeUrl + /lab
				notebookUrl = baseUrl + relativeUrl + "/lab"

				logger.InfoWithMetadata("Constructed full notebook URL for restart", map[string]interface{}{
					"build_id":           buildId,
					"original_url":       response.Url,
					"jupyter_api_url":    jupyterHubApiUrl,
					"base_url":           baseUrl,
					"relative_url":       relativeUrl,
					"final_notebook_url": notebookUrl,
				})
			}
			notebook.NotebookUrl = sql.NullString{String: notebookUrl, Valid: notebookUrl != ""}

			select {
			case <-done:
				// already closed
			default:
				close(done)
			}
		} else {
			notebook.Phase = "restarting"
		}

		// Use UpdateNotebookBuildStatus for consistency with build status endpoint
		if err := notebook.UpdateNotebookBuildStatus(app); err != nil {
			logger.ErrorWithMetadata("Error updating notebook build status during restart", map[string]interface{}{
				"build_id":    buildId,
				"phase":       notebook.Phase,
				"event_count": eventCount,
			}, err)
		}

		logger.DebugWithMetadata("Notebook restart status updated", map[string]interface{}{
			"build_id":    buildId,
			"phase":       notebook.Phase,
			"event_count": eventCount,
		})
	})

	logger.InfoWithMetadata("SSE subscription attempt completed", map[string]interface{}{
		"build_id":     buildId,
		"progress_url": progressUrl,
		"has_error":    sseserror != nil,
	})

	duration := time.Since(start)

	if sseserror != nil {
		logger.ErrorWithMetadata("SSE restart subscription failed", map[string]interface{}{
			"build_id":        buildId,
			"progress_url":    progressUrl,
			"duration":        duration.String(),
			"events_received": eventCount,
			"error_message":   sseserror.Error(),
		}, sseserror)
		return sseserror
	}

	logger.InfoWithMetadata("SSE subscription started successfully, waiting for events", map[string]interface{}{
		"build_id":     buildId,
		"progress_url": progressUrl,
	})

	// Wait for restart to complete or timeout
	select {
	case <-done:
		logger.InfoWithMetadata("SSE restart completed for buildId", map[string]interface{}{
			"build_id":        buildId,
			"events_received": eventCount,
		})
	case <-time.After(2 * time.Minute):
		logger.WarnWithMetadata("SSE restart timed out for buildId", map[string]interface{}{
			"build_id":        buildId,
			"events_received": eventCount,
			"progress_url":    progressUrl,
		})
		sseErr = fmt.Errorf("SSE restart timed out after 2 minutes, received %d events", eventCount)
	}

	duration = time.Since(start)

	if eventCount == 0 {
		logger.InfoWithMetadata("No SSE events received during restart", map[string]interface{}{
			"build_id":     buildId,
			"progress_url": progressUrl,
			"duration":     duration.String(),
		})
	}

	logger.InfoWithMetadata("SSE restart connection completed", map[string]interface{}{
		"build_id":        buildId,
		"duration":        duration.String(),
		"events_received": eventCount,
	})

	return sseErr
}

// RestartNotebookTaskHandler creates a task handler for restart notebook
func RestartNotebookTaskHandler(app *application.Application) func(context.Context, interface{}) error {
	return func(ctx context.Context, payload interface{}) error {
		start := time.Now()

		taskPayload, ok := payload.(*RestartNotebookTaskPayload)
		if !ok {
			logger.ErrorWithMetadata("Invalid payload type for restart notebook task", map[string]interface{}{
				"expected_type": "*RestartNotebookTaskPayload",
				"actual_type":   fmt.Sprintf("%T", payload),
			}, nil)
			return nil
		}

		logger.InfoWithMetadata("Processing restart notebook task", map[string]interface{}{
			"build_id":     taskPayload.BuildId,
			"user_id":      taskPayload.UserId,
			"progress_url": taskPayload.ProgressUrl,
		})

		// Check for context cancellation before starting
		select {
		case <-ctx.Done():
			logger.WarnWithMetadata("Restart task cancelled before execution", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"reason":   ctx.Err().Error(),
			})
			return ctx.Err()
		default:
		}

		err := handleRestartSse(app, taskPayload.ProgressUrl, taskPayload.BuildId, taskPayload.UserId)

		duration := time.Since(start)

		if err != nil {
			logger.ErrorWithMetadata("Restart notebook task failed", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"user_id":  taskPayload.UserId,
				"duration": duration.String(),
			}, err)
			return err
		}

		logger.InfoWithMetadata("Restart notebook task completed successfully", map[string]interface{}{
			"build_id": taskPayload.BuildId,
			"user_id":  taskPayload.UserId,
			"duration": duration.String(),
		})

		return nil
	}
}

// RegisterTask registers the restart notebook task with the task queue
func RegisterTask(app *application.Application) {
	logger.InfoWithMetadata("Restart notebook task handler registered", map[string]interface{}{
		"handler_type": "restart-notebook",
		"queue_type":   "in-memory",
	})
}

// AddRestartNotebookTask adds a restart notebook task to the queue
func AddRestartNotebookTask(ctx context.Context, app *application.Application, progressUrl, buildId string, userId int) error {
	logger.DebugWithMetadata("Adding restart notebook task to queue", map[string]interface{}{
		"build_id":     buildId,
		"user_id":      userId,
		"progress_url": progressUrl,
	})

	payload := &RestartNotebookTaskPayload{
		ProgressUrl: progressUrl,
		BuildId:     buildId,
		UserId:      userId,
	}

	err := app.TaskQueue.AddTask(ctx, "restart-notebook", payload, RestartNotebookTaskHandler(app))

	if err != nil {
		logger.ErrorWithMetadata("Failed to add restart notebook task to queue", map[string]interface{}{
			"build_id":     buildId,
			"user_id":      userId,
			"progress_url": progressUrl,
		}, err)
		return err
	}

	logger.InfoWithMetadata("Restart notebook task added to queue successfully", map[string]interface{}{
		"build_id":  buildId,
		"user_id":   userId,
		"task_type": "restart-notebook",
	})

	return nil
}
