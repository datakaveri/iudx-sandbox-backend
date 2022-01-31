package models

import (
	"database/sql"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type Notebook struct {
	UserId       int
	NotebookId   string         `json:"notebookId"`
	NotebookName string         `json:"name"`
	NotebookUrl  sql.NullString `json:"url"`
	RepoName     string         `json:"repoName"`
	BuildId      string         `json:"buildId"`
	Phase        string         `json:"phase"`
	Message      string         `json:"message"`
	Token        sql.NullString `json:"token"`
	ImageName    string         `json:"imageName"`
	CreatedAt    string         `json:"createdAt"`
	LastUsed     string         `json:"lastUsed"`
}

func (g *Notebook) Create(app *application.Application) error {
	stmt := `
		INSERT INTO notebook (
			"notebookId",
			"name",
			"buildId",
			"userId",
			"phase"
		)
		VALUES ($1, $2, $3, $4, $5);
	`

	result, err := app.DB.Client.Exec(stmt, g.NotebookId, g.NotebookName, g.BuildId, g.UserId, g.Phase)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("Notebook: rows inserted: %v\n", count)
	return nil
}

func (g *Notebook) List(app *application.Application) ([]Notebook, error) {
	stmt := `
		SELECT "notebookId", "name", "url", "buildId", "status", "createdAt", "lastUsed"
		FROM notebook;
	`

	rows, err := app.DB.Client.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notebooks := []Notebook{}
	for rows.Next() {
		var notebook Notebook
		rows.Scan(&notebook.NotebookId, &notebook.NotebookName, &notebook.NotebookUrl, &notebook.BuildId, &notebook.CreatedAt, &notebook.LastUsed)
		notebooks = append(notebooks, notebook)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return notebooks, nil
}

func (g *Notebook) Delete(app *application.Application, itemName string) error {
	stmt := `
		DELETE FROM notebook
		WHERE "notebookId" = $1;
	`

	result, err := app.DB.Client.Exec(stmt, itemName)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("Notebook: rows deleted: %v\n", count)
	return nil
}

func (g *Notebook) UpdateNotebookStatus(app *application.Application) error {
	stmt := `
		UPDATE notebook
		SET "phase" = $2, "message" = $3, "token" = $4, "imageName" = $5, "url" = $6
		WHERE "buildId"=$1;
	`

	result, err := app.DB.Client.Exec(stmt, g.BuildId, g.Phase, g.Message, g.Token, g.ImageName, g.NotebookUrl)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("Notebook: rows affected: %v\n", count)
	return nil
}

func (g *Notebook) GetBuildStatus(app *application.Application, buildId string) (Notebook, error) {
	stmt := `
		SELECT "url", "token", "phase"
		FROM notebook 
		WHERE "buildId" = $1;
	`

	notebook := Notebook{}

	row := app.DB.Client.QueryRow(stmt, buildId)

	err := row.Scan(&notebook.NotebookUrl, &notebook.Token, &notebook.Phase)
	if err != nil {
		return notebook, err
	}

	return notebook, nil
}
