package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)

type ReferenceResource struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	AdditionalInfoURL string `json:"additionalInfoURL"`
	DatasetID         string `json:"datasetID"`
}

type ReferenceResourceResponse struct {
	Id                string
	Name              string
	Description       string
	AdditionalInfoURL string
	DatasetID         string
}

func (g *ReferenceResource) ListReferenceResource(app *application.Application, unique_id string) ([]ReferenceResourceResponse, error) {
	stmt := `select * from referenceResources
	where "datasetID" = $1`

	rows, err := app.DB.Client.Query(stmt, unique_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	referenceResources := []ReferenceResourceResponse{}

	for rows.Next() {
		var referenceResource ReferenceResourceResponse
		err := rows.Scan(&referenceResource.Id, &referenceResource.Name,
			&referenceResource.Description, &referenceResource.AdditionalInfoURL,
			&referenceResource.DatasetID)

		if err != nil {
			return nil, err
		}

		referenceResources = append(referenceResources, referenceResource)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return referenceResources, nil
}

func (g *ReferenceResource) OnboardReferenceResource(app *application.Application) error {
	stmt := `
		insert into referenceResources (
			"id",
			"name",
			"description",
			"additionalInfoURL",
			"datasetID"
		) values ($1, $2, $3, $4, $5);
	`

	result, err := app.DB.Client.Exec(stmt, g.Id,
		g.Name,
		g.Description,
		g.AdditionalInfoURL,
		g.DatasetID)

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
