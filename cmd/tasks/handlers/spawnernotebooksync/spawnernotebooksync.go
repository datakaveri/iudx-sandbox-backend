package spawnernotebooksync

import (
	"encoding/json"
	"net/url"

	"github.com/iudx-sandbox-backend/cmd/api/models"
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/r3labs/sse/v2"
	"github.com/vmihailenco/taskq/v3"
)

type SpawnerNotebookSyncTask struct {
	Task *taskq.Task
}

func customHeader(cookie string) func(c *sse.Client) {
	return func(c *sse.Client) {
		c.Headers = map[string]string{
			"cookie": cookie,
		}
	}
}

func buildNotebookSse(app *application.Application, buildUrl, cookie, buildId string, userId int) {
	sseClient := sse.NewClient(buildUrl, customHeader(cookie))

	sseserror := sseClient.SubscribeRaw(func(msg *sse.Event) {
		notebook := &models.Notebook{}
		json.Unmarshal(msg.Data, notebook)
		notebook.BuildId = buildId

		if notebook.Phase == "ready" {
			// FIXME slightly inconsistent approach its unreliable maybe we can change spawner name but for single server that won't work
			parsedUrl, err := url.Parse(notebook.NotebookUrl)
			baseUrl := parsedUrl.Path + "/"
			if err != nil {
				logger.Error.Printf("Error parsing url %v\n", err)
			}
			spawner := &models.Spawner{}
			res, err := spawner.GetSpawnerIdBasedOnBaseUrl(app, baseUrl, userId)
			if err != nil {
				logger.Error.Printf("Error finding spawner id %v\n", err)
			}
			notebook.SpawnerId = res.Id
			if err := notebook.UpdateNotebookSpawnerId(app); err != nil {
				logger.Error.Printf("Error failed to update spawner id %v\n", err)
			}
		}

		if err := notebook.UpdateNotebookBuildStatus(app); err != nil {
			logger.Error.Printf("Binder: Error in building notebook %v\n", err)
			return
		}
	})

	if sseserror != nil {
		logger.Error.Printf("Error in sse %v\n", sseserror)
	}
}

func RegisterTask() *SpawnerNotebookSyncTask {
	var SpawnerIdSyncTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "spawnerIdSync",
		Handler: func(app *application.Application, buildUrl, cookie, buildId string, userId int) error {
			buildNotebookSse(app, buildUrl, cookie, buildId, userId)
			return nil
		},
	})

	return &SpawnerNotebookSyncTask{
		Task: SpawnerIdSyncTask,
	}
}
