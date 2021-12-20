package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type Repo struct {
	RepoName    string `json:"repoName"`
	Description string `json:"description"`
	GithubUrl   string `json:"githubUrl"`
}

func (g *Repo) Create(app *application.Application) error {
	stmt := `
		INSERT INTO gallery (
			"repoName",
			"description",
			"githubUrl"
		)
		VALUES ($1, $2, $3);
	`

	result, err := app.DB.Client.Exec(stmt, g.RepoName, g.Description, g.GithubUrl)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("Gallery: rows inserted: %v\n", count)
	return nil
}

func (g *Repo) Get(app *application.Application, repoName string) (Repo, error) {
	stmt := `
		SELECT "repoName", "description", "githubUrl"
		FROM gallery 
		WHERE "repoName" = $1;
	`

	gallery := Repo{}

	row := app.DB.Client.QueryRow(stmt, repoName)

	err := row.Scan(&gallery.RepoName, &gallery.Description, &gallery.GithubUrl)
	if err != nil {
		return gallery, err
	}

	return gallery, nil
}

func (g *Repo) List(app *application.Application) ([]Repo, error) {
	stmt := `
		SELECT "repoName", "description", "githubUrl"
		FROM gallery;
	`

	rows, err := app.DB.Client.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := []Repo{}
	for rows.Next() {
		var repo Repo
		rows.Scan(&repo.RepoName, &repo.Description, &repo.GithubUrl)
		repos = append(repos, repo)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (g *Repo) Delete(app *application.Application, itemName string) error {
	stmt := `
		DELETE FROM gallery
		WHERE "repoName" = $1;
	`

	result, err := app.DB.Client.Exec(stmt, itemName)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Info.Printf("Gallery: rows deleted: %v\n", count)
	return nil
}
