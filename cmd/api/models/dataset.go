package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/lib/pq"
)

type ResourceItem struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	AdditionalInfoURL string `json:"additionalInfoURL"`
}
type ReferenceResource struct {
	Items []ResourceItem `json:"items"`
}

type Dataset struct {
	Id                 string          `json:"id"`
	AccessPolicy       string          `json:"accessPolicy"`
	Description        string          `json:"description"`
	Domain             string          `json:"domain"`
	Icon               string          `json:"icon"`
	IUDXResourceAPIs   []string        `json:"iudxResourceAPIs"`
	Label              string          `json:"label"`
	Name               string          `json:"name"`
	Provider           json.RawMessage `json:"provider"`
	ReferenceResources json.RawMessage `json:"referenceResources"`
	RepositoryURL      string          `json:"repositoryURL"`
	Tags               []string        `json:"tags"`
	Type               []string        `json:"type"`
	Unique_id          string          `json:"unique_id"`
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
	ReferenceResources json.RawMessage
	RepositoryURL      string
	Tags               []string
	Type               []string
	Unique_id          string
}

func (r ReferenceResource) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (a *ReferenceResource) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
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
			&dataset.ReferenceResources,
			&dataset.RepositoryURL,
			pq.Array(&dataset.Tags),
			pq.Array(&dataset.Type), &dataset.Unique_id)

		if err != nil {
			return nil, err
		}

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
			"domain",
			"description",
			"icon",           
			"iudxResourceAPIs",
			"label",
			"name",
			"provider",
			"referenceResources",
			"repositoryURL",
			"tags",
			"type",
			"unique_id"
		) values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14
		);
	`

	result, err := app.DB.Client.Exec(stmt, g.Id, g.AccessPolicy,
		g.Description, g.Domain, g.Icon, pq.Array(g.IUDXResourceAPIs),
		g.Label, g.Name, g.Provider, g.ReferenceResources,
		g.RepositoryURL, pq.Array(g.Tags),
		pq.Array(g.Type), g.Unique_id)

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
			&dataset.ReferenceResources,
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
