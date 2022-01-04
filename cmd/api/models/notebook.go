package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type Notebook struct {
	NotebookId   string `json:"notebookId"`
	NotebookName string `json:"name"`
	NotebookUrl  string `json:"url"`
	RepoName     string `json:"repoName"`
	BuildId      string `json:"buildId"`
	BuildStatus  string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	LastUsed     string `json:"lastUsed"`
}

func (g *Notebook) Create(app *application.Application) error {
	stmt := `
		INSERT INTO notebook (
			"notebookId",
			"name",
			"buildId",
			"status",
		)
		VALUES ($1, $2, $3, $4);
	`

	result, err := app.DB.Client.Exec(stmt, g.NotebookId, g.NotebookName, g.BuildId, g.BuildStatus)
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
		SELECT "notebookId", "name", "url", "buildId", "status", "createdAt", "lastUsed",
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
		rows.Scan(&notebook.NotebookId, &notebook.NotebookName, &notebook.NotebookUrl, &notebook.BuildId, &notebook.BuildStatus, &notebook.CreatedAt, &notebook.LastUsed)
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

func (g *Notebook) UpdateNotebookReady(app *application.Application, buildId, status, url string) error {
	stmt := `
		UPDATE notebook
		SET status = $2, url = $3
		WHERE "buildId" = $1;
	`

	result, err := app.DB.Client.Exec(stmt, buildId, status, url)
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

func (g *Notebook) GetBuildStatus(app *application.Application, buildId string) (Notebook, error) {
	stmt := `
	SELECT "url"
	FROM notebook 
	WHERE "buildId" = $1;
`

	notebook := Notebook{}

	row := app.DB.Client.QueryRow(stmt, buildId)

	err := row.Scan(&notebook.NotebookUrl)
	if err != nil {
		return notebook, err
	}

	return notebook, nil
}
