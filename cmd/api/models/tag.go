package models

import "github.com/iudx-sandbox-backend/pkg/application"

type Tag struct {
	Tag string `json:"tag"`
}

func (g *Tag) ListTags(app *application.Application) ([]string, error) {
	stmt := `
		select distinct(unnest(tags)) from dataset;
	`

	rows, err := app.DB.Client.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tags := []string{}

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)

		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return tags, nil

}
