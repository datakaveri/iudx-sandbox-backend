package models

import (
	"time"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/lib/pq"
)

type Resource struct {
	Id            string    `json:"id"`
	CreatedAt     string    `json:"createdAt"`
	Dataset       string    `json:"dataset"`
	Description   string    `json:"description"`
	DownloadURL   string    `json:"downloadURL"`
	Icon          string    `json:"icon"`
	Instance      string    `json:"instance"`
	ItemCreatedAt string    `json:"itemCreatedAt"`
	ItemStatus    string    `json:"itemStatus"`
	Label         string    `json:"label"`
	Name          string    `json:"name"`
	Provider      string    `json:"provider"`
	ResourceGroup string    `json:"resourceGroup"`
	Tags          []string  `json:"tags"`
	Type          []string  `json:"type"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ResourceResponse struct {
	Id            string
	CreatedAt     string
	Dataset       string
	Description   string
	DownloadURL   string
	Icon          string
	Instance      string
	ItemCreatedAt string
	ItemStatus    string
	Label         string
	Name          string
	Provider      string
	ResourceGroup string
	Tags          []string
	Type          []string
	UpdatedAt     time.Time
}

func (g *Resource) ListResource(app *application.Application, unique_id string) ([]ResourceResponse, error) {
	stmt := `
		select * from resource
		where "resourceGroup" = (
			select "id" from dataset
			where "unique_id" = $1
		);
	`

	rows, err := app.DB.Client.Query(stmt, unique_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resources := []ResourceResponse{}

	for rows.Next() {
		var resource ResourceResponse
		err := rows.Scan(
			&resource.Id, &resource.CreatedAt, &resource.Dataset, &resource.Description,
			&resource.DownloadURL, &resource.Icon, &resource.Instance,
			&resource.ItemCreatedAt, &resource.ItemStatus, &resource.Label, &resource.Name,
			&resource.Provider, &resource.ResourceGroup, pq.Array(&resource.Tags),
			pq.Array(&resource.Type), &resource.UpdatedAt)

		if err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (g *Resource) OnboardResource(app *application.Application) error {
	stmt := `
		insert into resource (
			"id",
			"createdAt",
			"dataset",
			"description",
			"downloadURL",
			"icon",
			"instance",
			"itemCreatedAt",
			"itemStatus",
			"label",
			"name",
			"provider",
			"resourceGroup",
			"tags",
			"type",
			"updatedAt"
		) values (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16
		);
	`

	result, err := app.DB.Client.Exec(stmt, g.Id, g.CreatedAt, g.Dataset, g.Description,
		g.DownloadURL, g.Icon, g.Instance, g.ItemCreatedAt, g.ItemStatus, g.Label,
		g.Name, g.Provider, g.ResourceGroup, pq.Array(g.Tags), pq.Array(g.Type),
		g.UpdatedAt)

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
