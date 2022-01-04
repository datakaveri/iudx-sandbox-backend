package models

import (
	"encoding/json"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type BuildLog struct {
	BuildId     string       `json:"buildId"`
	Phase       string       `json:"phase"`
	Message     string       `json:"message"`
	Token       string       `json:"token"`
	Progress    json.Decoder `json:"progress"`
	ImageName   string       `json:"imageName"`
	NotebookUrl string       `json:"url"`
}

func (g *BuildLog) Create(app *application.Application) error {
	stmt := `
		INSERT INTO buildlog (
			"buildId",
			"phase",
			"message",
			"token",
			"progress",
			"imageName",
			"url",
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	result, err := app.DB.Client.Exec(stmt, g.BuildId, g.Phase, g.Message, g.Token, g.Progress, g.ImageName, g.NotebookUrl)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("BuildLog: rows inserted: %v\n", count)
	return nil
}
