package models

import (
	"encoding/json"
	"time"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/lib/pq"
)

// type ReferenceResource struct {
// 	Id                string `json:"id"`
// 	Name              string `json:"name"`
// 	Description       string `json:"description"`
// 	AdditionalInfoURL string `json:"additionalInfoURL"`
// }

// type ReferenceResources map[string]interface{}

type Dataset struct {
	Id               string          `json:"id"`
	AccessPolicy     string          `json:"accessPolicy"`
	CreatedAt        time.Time       `json:"createdAt"`
	Description      string          `json:"description"`
	Icon             string          `json:"icon"`
	Instance         string          `json:"instance"`
	ItemCreatedAt    time.Time       `json:"itemCreatedAt"`
	ItemStatus       string          `json:"itemStatus"`
	IUDXResourceAPIs []string        `json:"iudxResourceAPIs"`
	Label            string          `json:"label"`
	Location         json.RawMessage `json:"location"`
	Name             string          `json:"name"`
	Provider         json.RawMessage `json:"provider"`
	// ReferenceResources ReferenceResource `json:"referenceResources"`
	RepositoryURL  string `json:"repositoryURL"`
	ResourceServer string `json:"resourceServer"`
	ResourceType   string `json:"resourceType"`
	Resources      int
	Schema         string    `json:"schema"`
	Tags           []string  `json:"tags"`
	Type           []string  `json:"type"`
	Unique_id      string    `json:"unique_id"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Views          int
}

type DatasetResponse struct {
	Id               string
	AccessPolicy     string
	CreatedAt        time.Time
	Description      string
	Icon             string
	Instance         string
	ItemCreatedAt    string
	ItemStatus       string
	IUDXResourceAPIs []string
	Label            string
	Location         json.RawMessage
	Name             string
	Provider         json.RawMessage
	// ReferenceResources []ReferenceResource
	RepositoryURL  string
	ResourceServer string
	ResourceType   string
	Resources      int
	Schema         string
	Tags           []string
	Type           []string
	Unique_id      string
	UpdatedAt      time.Time
	Views          int
}

// func (r ReferenceResources) Value() (driver.Value, error) {
// 	return json.Marshal(r)
// }

// func (a *ReferenceResources) Scan(value interface{}) error {
// 	b, ok := value.([]byte)
// 	if !ok {
// 		return errors.New("type assertion to []byte failed")
// 	}

// 	return json.Unmarshal(b, &a)
// }

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
			&dataset.Id, &dataset.AccessPolicy, &dataset.CreatedAt,
			&dataset.Description, &dataset.Icon, &dataset.Instance,
			&dataset.ItemCreatedAt, &dataset.ItemStatus,
			pq.Array(&dataset.IUDXResourceAPIs), &dataset.Label,
			&dataset.Location, &dataset.Name, &dataset.Provider,
			// pq.Array(&dataset.ReferenceResources),
			&dataset.RepositoryURL, &dataset.ResourceServer,
			&dataset.ResourceType, &dataset.Resources,
			&dataset.Schema, pq.Array(&dataset.Tags),
			pq.Array(&dataset.Type), &dataset.Unique_id,
			&dataset.UpdatedAt, &dataset.Views)

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
			"createdAt",
			"description",
			"icon",           
			"instance",
			"itemCreatedAt",
			"itemStatus",
			"iudxResourceAPIs",
			"label",
			"location",
			"name",
			"provider"
			"repositoryURL",
			"resourceServer",
			"resourceType",
			"resources",
			"schema",
			"tags",
			"type",
			"unique_id",
			"updatedAt",
			"views"
		) values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23
		);
	`

	result, err := app.DB.Client.Exec(stmt, g.Id, g.AccessPolicy, g.CreatedAt,
		g.Description, g.Icon, g.Instance,
		g.ItemCreatedAt, g.ItemStatus, pq.Array(g.IUDXResourceAPIs),
		g.Label, g.Location, g.Name, g.Provider,
		g.RepositoryURL, g.ResourceServer, g.ResourceType,
		g.Resources, g.Schema, pq.Array(g.Tags),
		pq.Array(g.Type), g.Unique_id, g.UpdatedAt, g.Views)

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
			&dataset.Id, &dataset.AccessPolicy, &dataset.CreatedAt,
			&dataset.Description, &dataset.Icon, &dataset.Instance,
			&dataset.ItemCreatedAt, &dataset.ItemStatus,
			pq.Array(&dataset.IUDXResourceAPIs), &dataset.Label,
			&dataset.Location, &dataset.Name, &dataset.Provider,
			// pq.Array(&dataset.ReferenceResources),
			&dataset.RepositoryURL, &dataset.ResourceServer,
			&dataset.ResourceType, &dataset.Resources,
			&dataset.Schema, pq.Array(&dataset.Tags),
			pq.Array(&dataset.Type), &dataset.Unique_id,
			&dataset.UpdatedAt, &dataset.Views)

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
