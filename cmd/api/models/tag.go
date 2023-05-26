package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
)

type Tag struct {
	Tag string
}

func (g *Tag) ListTags(app *application.Application) ([]string, error) {
	stmt := `
		select distinct(unnest(tags)) as tags from dataset order by tags asc;
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

func (g *Tag) ListDomains(app *application.Application) ([]string, error) {
	stmt := `
		select distinct("domain") as domains from dataset order by domains asc;
	`

	rows, err := app.DB.Client.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	domains := []string{}

	for rows.Next() {
		var domain string
		err := rows.Scan(&domain)

		if err != nil {
			return nil, err
		}

		domains = append(domains, domain)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return domains, nil

}
