package spawnernotebooksync

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/r3labs/sse/v2"
)

type SpawnerNotebookSyncTaskPayload struct {
	BuildUrl string `json:"build_url"`
	Cookie   string `json:"cookie"`
	BuildId  string `json:"build_id"`
	UserId   int    `json:"user_id"`
}

func customHeader(cookie string) func(c *sse.Client) {
	return func(c *sse.Client) {
		c.Headers = map[string]string{
			"cookie": cookie,
		}
	}
}

func buildNotebookSse(app *application.Application, buildUrl, cookie, buildId string, userId int) error {
	start := time.Now()

	logger.InfoWithMetadata("Starting SSE connection for notebook build", map[string]interface{}{
		"build_id":      buildId,
		"user_id":       userId,
		"build_url":     buildUrl,
		"has_cookie":    cookie != "",
		"cookie_length": len(cookie),
	})

	logger.InfoWithMetadata("SSE connection parameters", map[string]interface{}{
		"build_id":     buildId,
		"build_url":    buildUrl,
		"cookie_value": cookie,
		"user_id":      userId,
	})

	if buildUrl == "" {
		err := fmt.Errorf("build URL is empty")
		logger.ErrorWithMetadata("SSE connection failed - missing build URL", map[string]interface{}{
			"build_id": buildId,
			"user_id":  userId,
		}, err)
		return err
	}

	if cookie == "" {
		logger.WarnWithMetadata("SSE connection starting without cookie", map[string]interface{}{
			"build_id":  buildId,
			"build_url": buildUrl,
		})
	}

	sseClient := sse.NewClient(buildUrl, customHeader(cookie))

	// Disable TLS certificate verification to handle self-signed or invalid certificates
	sseClient.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	logger.InfoWithMetadata("SSE client created with TLS verification disabled", map[string]interface{}{
		"build_id":  buildId,
		"build_url": buildUrl,
	})

	eventCount := 0
	lastEventTime := time.Now()
	done := make(chan struct{})
	var sseErr error

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		eventCount++
		eventTime := time.Now()
		timeSinceLastEvent := eventTime.Sub(lastEventTime)
		lastEventTime = eventTime

		logger.InfoWithMetadata("SSE event received", map[string]interface{}{
			"build_id":              buildId,
			"event_count":           eventCount,
			"time_since_last_event": timeSinceLastEvent.String(),
			"event_data_size":       len(msg.Data),
			"event_type":            msg.Event,
			"event_id":              msg.ID,
		})

		logger.InfoWithMetadata("Raw SSE event data", map[string]interface{}{
			"build_id":   buildId,
			"raw_data":   string(msg.Data),
			"event_type": msg.Event,
		})

		// Temporary struct for unmarshaling SSE event with string fields
		var eventData struct {
			Phase   string `json:"phase"`
			Message string `json:"message"`
			Image   string `json:"image"`
			Token   string `json:"token"`
			Url     string `json:"url"`
		}

		if err := json.Unmarshal(msg.Data, &eventData); err != nil {
			logger.ErrorWithMetadata("Failed to unmarshal SSE event data", map[string]interface{}{
				"build_id":    buildId,
				"event_count": eventCount,
				"raw_data":    string(msg.Data),
				"data_size":   len(msg.Data),
			}, err)
			return
		}

		// Create notebook and map fields from eventData
		notebook := &models.Notebook{}
		notebook.BuildId = buildId
		notebook.Phase = eventData.Phase
		notebook.Message = eventData.Message
		notebook.ImageName = eventData.Image
		notebook.Token = sql.NullString{String: eventData.Token, Valid: eventData.Token != ""}

		// Append /lab to the URL, ensuring no double slashes
		notebookUrl := eventData.Url
		if notebookUrl != "" {
			// Remove trailing slash if present
			notebookUrl = strings.TrimSuffix(notebookUrl, "/")
			// Append /lab
			notebookUrl = notebookUrl + "/lab"
		}
		notebook.NotebookUrl = sql.NullString{String: notebookUrl, Valid: notebookUrl != ""}

		logger.InfoWithMetadata("Notebook status update received", map[string]interface{}{
			"build_id":     buildId,
			"phase":        notebook.Phase,
			"notebook_url": notebook.NotebookUrl.String,
			"event_count":  eventCount,
			"ready":        notebook.Phase == "ready",
		})

		if notebook.Phase == "ready" || notebook.Phase == "failed" {
			select {
			case <-done:
				// already closed
			default:
				close(done)
			}
		}

		if notebook.Phase == "ready" {
			logger.InfoWithMetadata("Notebook build completed - processing spawner ID", map[string]interface{}{
				"build_id":     buildId,
				"notebook_url": notebook.NotebookUrl.String,
			})

			parsedUrl, err := url.Parse(notebook.NotebookUrl.String)
			if err != nil {
				logger.ErrorWithMetadata("Failed to parse notebook URL for spawner ID extraction", map[string]interface{}{
					"build_id":     buildId,
					"notebook_url": notebook.NotebookUrl.String,
				}, err)
				return
			}

			baseUrl := parsedUrl.Path + "/"
			// Remove /lab/ from base URL for spawner lookup since JupyterHub stores base URLs without /lab
			baseUrl = strings.TrimSuffix(baseUrl, "/lab/") + "/"
			logger.DebugWithMetadata("Extracted base URL for spawner lookup", map[string]interface{}{
				"build_id": buildId,
				"base_url": baseUrl,
				"user_id":  userId,
			})

			spawner := &models.Spawner{}
			res, err := spawner.GetSpawnerIdBasedOnBaseUrl(app, baseUrl, userId)
			if err != nil {
				logger.ErrorWithMetadata("Failed to find spawner ID based on base URL", map[string]interface{}{
					"build_id": buildId,
					"base_url": baseUrl,
					"user_id":  userId,
				}, err)
				return
			}

			notebook.SpawnerId = res.Id
			logger.InfoWithMetadata("Spawner ID found and assigned", map[string]interface{}{
				"build_id":   buildId,
				"spawner_id": notebook.SpawnerId,
				"base_url":   baseUrl,
			})

			if err := notebook.UpdateNotebookSpawnerId(app); err != nil {
				logger.ErrorWithMetadata("Failed to update notebook spawner ID", map[string]interface{}{
					"build_id":   buildId,
					"spawner_id": notebook.SpawnerId,
				}, err)
				return
			}

			logger.InfoWithMetadata("Notebook spawner ID updated successfully", map[string]interface{}{
				"build_id":   buildId,
				"spawner_id": notebook.SpawnerId,
			})
		}

		if err := notebook.UpdateNotebookBuildStatus(app); err != nil {
			logger.ErrorWithMetadata("Failed to update notebook build status", map[string]interface{}{
				"build_id":    buildId,
				"phase":       notebook.Phase,
				"event_count": eventCount,
			}, err)
			return
		}

		logger.DebugWithMetadata("Notebook build status updated", map[string]interface{}{
			"build_id":    buildId,
			"phase":       notebook.Phase,
			"event_count": eventCount,
		})
	})

	duration := time.Since(start)

	if sseserror != nil {
		logger.ErrorWithMetadata("SSE subscription failed", map[string]interface{}{
			"build_id":        buildId,
			"build_url":       buildUrl,
			"duration":        duration.String(),
			"events_received": eventCount,
		}, sseserror)
		return sseserror
	}

	// Wait for build to complete or timeout
	select {
	case <-done:
		logger.InfoWithMetadata("SSE build completed for buildId", map[string]interface{}{
			"build_id": buildId,
		})
	case <-time.After(2 * time.Minute):
		logger.WarnWithMetadata("SSE build timed out for buildId", map[string]interface{}{
			"build_id": buildId,
		})
		sseErr = fmt.Errorf("SSE build timed out")
	}

	duration = time.Since(start)

	logger.InfoWithMetadata("SSE connection completed", map[string]interface{}{
		"build_id":        buildId,
		"duration":        duration.String(),
		"events_received": eventCount,
	})

	return sseErr
}

// SpawnerNotebookSyncTaskHandler creates a task handler for spawner notebook sync
func SpawnerNotebookSyncTaskHandler(app *application.Application) func(context.Context, interface{}) error {
	return func(ctx context.Context, payload interface{}) error {
		start := time.Now()

		taskPayload, ok := payload.(*SpawnerNotebookSyncTaskPayload)
		if !ok {
			logger.ErrorWithMetadata("Invalid payload type for spawner notebook sync task", map[string]interface{}{
				"expected_type": "*SpawnerNotebookSyncTaskPayload",
				"actual_type":   fmt.Sprintf("%T", payload),
			}, nil)
			return nil
		}

		logger.InfoWithMetadata("Processing spawner notebook sync task", map[string]interface{}{
			"build_id":   taskPayload.BuildId,
			"user_id":    taskPayload.UserId,
			"build_url":  taskPayload.BuildUrl,
			"has_cookie": taskPayload.Cookie != "",
		})

		// Check for context cancellation before starting
		select {
		case <-ctx.Done():
			logger.WarnWithMetadata("Task cancelled before execution", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"reason":   ctx.Err().Error(),
			})
			return ctx.Err()
		default:
		}

		err := buildNotebookSse(app, taskPayload.BuildUrl, taskPayload.Cookie,
			taskPayload.BuildId, taskPayload.UserId)

		duration := time.Since(start)

		if err != nil {
			logger.ErrorWithMetadata("Spawner notebook sync task failed", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"user_id":  taskPayload.UserId,
				"duration": duration.String(),
			}, err)
			return err
		}

		// Log performance metrics
		if duration > 5*time.Minute {
			logger.WarnWithMetadata("Spawner sync task completed but was very slow", map[string]interface{}{
				"build_id":  taskPayload.BuildId,
				"duration":  duration.String(),
				"threshold": "5m",
			})
		} else if duration > 2*time.Minute {
			logger.InfoWithMetadata("Spawner sync task completed (slow)", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"duration": duration.String(),
			})
		} else {
			logger.DebugWithMetadata("Spawner sync task completed efficiently", map[string]interface{}{
				"build_id": taskPayload.BuildId,
				"duration": duration.String(),
			})
		}

		return nil
	}
}

// RegisterTask registers the spawner notebook sync task with the task queue
func RegisterTask(app *application.Application) {
	logger.InfoWithMetadata("Spawner notebook sync task handler registered", map[string]interface{}{
		"handler_type": "spawner-notebook-sync",
		"queue_type":   "in-memory",
	})
}

// AddSpawnerSyncTask adds a spawner sync task to the queue
func AddSpawnerSyncTask(ctx context.Context, app *application.Application, buildUrl, cookie, buildId string, userId int) error {
	logger.DebugWithMetadata("Adding spawner sync task to queue", map[string]interface{}{
		"build_id":   buildId,
		"user_id":    userId,
		"build_url":  buildUrl,
		"has_cookie": cookie != "",
	})

	payload := &SpawnerNotebookSyncTaskPayload{
		BuildUrl: buildUrl,
		Cookie:   cookie,
		BuildId:  buildId,
		UserId:   userId,
	}

	err := app.TaskQueue.AddTask(ctx, "spawner-notebook-sync", payload, SpawnerNotebookSyncTaskHandler(app))

	if err != nil {
		logger.ErrorWithMetadata("Failed to add spawner sync task to queue", map[string]interface{}{
			"build_id":  buildId,
			"user_id":   userId,
			"build_url": buildUrl,
		}, err)
		return err
	}

	logger.InfoWithMetadata("Spawner sync task added to queue successfully", map[string]interface{}{
		"build_id":  buildId,
		"user_id":   userId,
		"task_type": "spawner-notebook-sync",
	})

	return nil
}
