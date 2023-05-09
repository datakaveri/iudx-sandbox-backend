package models

import (
	"encoding/json"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/lib/pq"
)

type ReferenceResource struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	AdditionalInfoURL string `json:"additionalInfoURL"`
}

type Dataset struct {
	Id                 string              `json:"id"`
	AccessPolicy       string              `json:"accessPolicy"`
	Description        string              `json:"description"`
	Domain             string              `json:"domain"`
	Icon               string              `json:"icon"`
	IUDXResourceAPIs   []string            `json:"iudxResourceAPIs"`
	Label              string              `json:"label"`
	Name               string              `json:"name"`
	Provider           json.RawMessage     `json:"provider"`
	ReferenceResources []ReferenceResource `json:"referenceResources"`
	RepositoryURL      string              `json:"repositoryURL"`
	Tags               []string            `json:"tags"`
	Type               []string            `json:"type"`
	Unique_id          string              `json:"unique_id"`
}

type ReferenceResourceResponse struct {
	Id                string
	Name              string
	Description       string
	AdditionalInfoURL string
	DatasetID         string
}
type DatasetResponse struct {
	Id                 string
	AccessPolicy       string
	Description        string
	Domain             string
	Icon               string
	IUDXResourceAPIs   []string
	Label              string
	Name               string
	Provider           json.RawMessage
	ReferenceResources []ReferenceResourceResponse
	RepositoryURL      string
	Tags               []string
	Type               []string
	Unique_id          string
}

func (g *Dataset) ListDataset(app *application.Application) ([]DatasetResponse, error) {
	stmt := `
		SELECT * from dataset;
	`
	rows, err := app.DB.Client.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	datasets := []DatasetResponse{}

	for rows.Next() {
		var dataset DatasetResponse
		err := rows.Scan(
			&dataset.Id, &dataset.AccessPolicy,
			&dataset.Description, &dataset.Domain, &dataset.Icon,
			pq.Array(&dataset.IUDXResourceAPIs), &dataset.Label,
			&dataset.Name, &dataset.Provider,
			&dataset.RepositoryURL,
			pq.Array(&dataset.Tags),
			pq.Array(&dataset.Type), &dataset.Unique_id)

		if err != nil {
			return nil, err
		}

		stmt2 := `
			select * from referenceResources
			where "datasetID" = $1;
		`

		rows2, err2 := app.DB.Client.Query(stmt2, &dataset.Id)

		if err2 != nil {
			return nil, err2
		}

		defer rows2.Close()

		referenceResources := []ReferenceResourceResponse{}

		for rows2.Next() {
			var reference ReferenceResourceResponse
			err2 := rows2.Scan(&reference.Id, &reference.Name,
				&reference.Description, &reference.AdditionalInfoURL,
				&reference.DatasetID)

			if err2 != nil {
				return nil, err2
			}

			referenceResources = append(referenceResources, reference)
		}

		dataset.ReferenceResources = referenceResources

		datasets = append(datasets, dataset)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return datasets, nil
}

func (g *Dataset) OnboardDataset(app *application.Application) error {

	stmt := `
		INSERT INTO dataset (
			"id",
			"accessPolicy",
			"description",
			"domain",
			"icon",
			"iudxResourceAPIs",
			"label",
			"name",
			"provider",
			"repositoryURL",
			"tags",
			"type",
			"unique_id"
		) values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13
		);
	`
	stmt2 := `
		INSERT INTO referenceResources (
			"id",
			"name",
			"description",
			"additionalInfoURL",
			"datasetID"
		) values (
			$1, $2, $3,
			$4, $5
		);
	`

	result, err := app.DB.Client.Exec(stmt, g.Id, g.AccessPolicy,
		g.Description, g.Domain, g.Icon, pq.Array(g.IUDXResourceAPIs),
		g.Label, g.Name, g.Provider,
		g.RepositoryURL, pq.Array(g.Tags),
		pq.Array(g.Type), g.Unique_id)

	for _, element := range g.ReferenceResources {
		_, err2 := app.DB.Client.Exec(stmt2, element.Id, element.Name,
			element.Description, element.AdditionalInfoURL, g.Id)

		if err2 != nil {
			return err2
		}
	}

	if err != nil {
		return err
	}

	count, err := result.RowsAffected()

	if err != nil {
		return nil
	}
	logger.Info.Printf("Dataset: rows inserted: %v\n", count)
	return nil
}

func (g *Dataset) GetDataset(app *application.Application, unique_id string) (DatasetResponse, error) {

	stmt := `
		SELECT * FROM dataset
		where "unique_id"=$1;
	`

	rows, err := app.DB.Client.Query(stmt, unique_id)

	dataset := DatasetResponse{}
	if err != nil {
		return dataset, err
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&dataset.Id, &dataset.AccessPolicy,
			&dataset.Description, &dataset.Domain, &dataset.Icon,
			pq.Array(&dataset.IUDXResourceAPIs), &dataset.Label,
			&dataset.Name, &dataset.Provider,
			&dataset.RepositoryURL, pq.Array(&dataset.Tags),
			pq.Array(&dataset.Type), &dataset.Unique_id)

		if err != nil {
			return dataset, err
		}
	}

	err = rows.Err()

	if err != nil {
		return dataset, err
	}

	return dataset, nil
}
