package models

import (
	"time"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/lib/pq"
)

type Dataset struct {
	AccessPolicy   string    `json:"accessPolicy"`
	CreatedAt      time.Time `json:"createdAt"`
	Description    string    `json:"description"`
	Icon           string    `json:"icon"`
	Id             string    `json:"id"`
	InstanceName   string    `json:"instanceName"`
	ItemCreatedAt  time.Time `json:"itemCreatedAt"`
	ItemStatus     string    `json:"itemStatus"`
	LabelTag       string    `json:"labelTag"`
	DatasetName    string    `json:"datasetName"`
	RepositoryURL  string    `json:"repositoryURL"`
	ResourceServer string    `json:"resourceServer"`
	ResourceType   string    `json:"resourceType"`
	Resources      int
	SchemaName     string   `json:"schemaName"`
	Tags           []string `json:"tags"`
}

type DatasetResponse struct {
	AccessPolicy   string
	CreatedAt      time.Time
	Description    string
	Icon           string
	Id             string
	InstanceName   string
	ItemCreatedAt  string
	ItemStatus     string
	LabelTag       string
	DatasetName    string
	RepositoryURL  string
	ResourceServer string
	ResourceType   string
	Resources      int
	SchemaName     string
	Tags           []string
}

func (g *Dataset) Get(app *application.Application) ([]DatasetResponse, error) {
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
			&dataset.AccessPolicy, &dataset.CreatedAt, &dataset.Description,
			&dataset.Icon, &dataset.Id, &dataset.InstanceName,
			&dataset.ItemCreatedAt, &dataset.ItemStatus, &dataset.LabelTag,
			&dataset.DatasetName, &dataset.RepositoryURL,
			&dataset.ResourceServer, &dataset.ResourceType, &dataset.Resources,
			&dataset.SchemaName, pq.Array(&dataset.Tags))

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

func (g *Dataset) Onboard(app *application.Application) error {
	stmt := `
		INSERT INTO dataset (
			"accessPolicy",
			"createdAt",
			"description",
			"icon",           
			"id", 
			"instanceName",
			"itemCreatedAt",
			"itemStatus",
			"labelTag",
			"datasetName",
			"repositoryURL",
			"resourceServer",
			"resourceType",
			"resources",
			"schemaName",
			"tags"
		) values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16
		);
	`

	result, err := app.DB.Client.Exec(stmt, g.AccessPolicy, g.CreatedAt,
		g.Description, g.Icon, g.Id, g.InstanceName, g.ItemCreatedAt, g.ItemStatus,
		g.LabelTag, g.DatasetName, g.RepositoryURL, g.ResourceServer, g.ResourceType,
		g.Resources, g.SchemaName, pq.Array(g.Tags))

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
