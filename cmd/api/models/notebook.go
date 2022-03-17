package models

import (
	"database/sql"
	"fmt"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type Notebook struct {
	UserId       int
	SpawnerId    int
	NotebookId   string `json:"notebookId"`
	NotebookName string `json:"name"`
	NotebookUrl  string `json:"url"`
	RepoName     string `json:"repoName"`
	BuildId      string `json:"buildId"`
	Phase        string `json:"phase"`
	Message      string `json:"message"`
	Token        string `json:"token"`
	ImageName    string `json:"imageName"`
	CreatedAt    string `json:"createdAt"`
	LastUsed     string `json:"lastUsed"`
}

type BuildStatusResponse struct {
	NotebookUrl sql.NullString `json:"url"`
	BuildId     string         `json:"buildId"`
	Phase       string         `json:"phase"`
	Token       sql.NullString `json:"token"`
	Message     sql.NullString `json:"message"`
}

type NotebookResponse struct {
	UserId       int
	ServerId     sql.NullInt64
	SpawnerId    sql.NullInt64
	SpawnerName  string
	NotebookId   string
	NotebookName string
	NotebookUrl  string
	RepoName     string
	Token        string
	Status       string
	BuildId      string
	CreatedAt    string
	LastUsed     string
}

func (g *Notebook) Create(app *application.Application) error {
	stmt := `
		INSERT INTO notebook (
			"notebookId",
			"name",
			"buildId",
			"userId",
			"phase",
			"repoName"
		)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	result, err := app.DB.Client.Exec(stmt, g.NotebookId, g.NotebookName, g.BuildId, g.UserId, g.Phase, g.RepoName)
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

func (g *Notebook) Get(app *application.Application, notebookId string) (Notebook, error) {
	stmt := `
		SELECT "buildId", "notebookId", "spawnerId", "userId", "url"
		FROM notebook 
		WHERE "notebookId" = $1;
	`

	notebook := Notebook{}

	row := app.DB.Client.QueryRow(stmt, notebookId)

	err := row.Scan(&notebook.BuildId, &notebook.NotebookId, &notebook.SpawnerId, &notebook.UserId, &notebook.NotebookUrl)
	if err != nil {
		return notebook, err
	}

	return notebook, nil
}

func (g *Notebook) List(app *application.Application, userId int) ([]NotebookResponse, error) {
	stmt := `
		SELECT users."id", spawners."server_id", spawners."id", spawners."name", notebook."notebookId", notebook."name", notebook."repoName", notebook."url", notebook."token", notebook."buildId", notebook."createdAt", spawners."last_activity"
		FROM spawners
		LEFT JOIN users
		ON users."id" = spawners."user_id"
		LEFT JOIN notebook
		ON notebook."userId" = users."id" AND notebook."spawnerId" = spawners."id"
		WHERE notebook."userId" = $1;
	`

	rows, err := app.DB.Client.Query(stmt, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notebooks := []NotebookResponse{}
	for rows.Next() {
		var notebook NotebookResponse
		err := rows.Scan(&notebook.UserId, &notebook.ServerId, &notebook.SpawnerId, &notebook.SpawnerName, &notebook.NotebookId, &notebook.NotebookName, &notebook.RepoName, &notebook.NotebookUrl, &notebook.Token, &notebook.BuildId, &notebook.CreatedAt, &notebook.LastUsed)
		if err != nil {
			return nil, err
		}
		fmt.Println(notebook)
		if notebook.ServerId.Int64 > 0 {
			notebook.Status = "RUNNING"
		} else {
			notebook.Status = "NOT RUNNING"
		}
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

func (g *Notebook) UpdateNotebookBuildStatus(app *application.Application) error {
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

func (g *Notebook) UpdateNotebookSpawnerId(app *application.Application) error {
	stmt := `
		UPDATE notebook
		SET "spawnerId" = $2
		WHERE "buildId"=$1;
	`

	result, err := app.DB.Client.Exec(stmt, g.BuildId, g.SpawnerId)
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

func (g *Notebook) GetBuildStatus(app *application.Application, buildId string) (BuildStatusResponse, error) {
	stmt := `
		SELECT "url", "token", "phase", "message"
		FROM notebook 
		WHERE "buildId" = $1;
	`

	response := BuildStatusResponse{}

	row := app.DB.Client.QueryRow(stmt, buildId)

	err := row.Scan(&response.NotebookUrl, &response.Token, &response.Phase, &response.Message)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (g *Notebook) UpdateNotebookStatus(app *application.Application) error {
	stmt := `
		UPDATE notebook
		SET "phase" = $2
		WHERE "notebookId"=$1;
	`

	result, err := app.DB.Client.Exec(stmt, g.NotebookId, g.Phase)
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

func (g *Notebook) GetNotebookIdByRepoName(app *application.Application, repoName string, userId int) (string, error) {
	stmt := `
		SELECT "notebookId"
		FROM notebook
		WHERE notebook."repoName" = $1 AND notebook."userId" = $2 AND notebook."phase" != "failed";
	`

	notebook := NotebookResponse{}

	result := app.DB.Client.QueryRow(stmt, repoName, userId)

	err := result.Scan(&notebook.NotebookId)
	if err != nil {
		return "", err
	}

	return notebook.NotebookId, nil
}

func (g *Notebook) GetSpawnerName(app *application.Application, userId int, notebookId string) (string, error) {
	stmt := `
		SELECT spawners."name"
		FROM spawners
		LEFT JOIN users
		ON users."id" = spawners."user_id"
		LEFT JOIN notebook
		ON notebook."userId" = users."id" AND notebook."spawnerId" = spawners."id"
		WHERE notebook."userId" = $1 AND notebook."notebookId" =$2;
	`

	notebook := NotebookResponse{}
	result := app.DB.Client.QueryRow(stmt, userId, notebookId)

	err := result.Scan(&notebook.RepoName)
	if err != nil {
		return "", err
	}

	return notebook.RepoName, nil
}

func (g *Notebook) RemoveSpawnerId(app *application.Application) error {
	stmt := `
		UPDATE notebook
		SET "spawnerId" = NULL
		WHERE "notebookId"=$1;
	`

	result, err := app.DB.Client.Exec(stmt, g.NotebookId)
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
