package models

import (
	"fmt"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/logger"
)



type Dataset struct {
	// accessPolicy	string
	// createdAt		time.Time
	description		string `json:"description"`
	// icon			string
	// id				string
	// instance		string
	// itemCreatedAt	string
	// itemStatus		string
	label			string `json:"label"`
	name			string `json:"name"`
	// repositoryURL	string
	// resourceServer	string
	// resourceType	string
	// resources		int
	// schema			string
	// tags			[]string	
}

func (g *Dataset) Get(app *application.Application) ([]Dataset, error) {
	stmt := `
		SELECT * from dataset;
	`
	rows, err := app.DB.Client.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	datasets := []Dataset{}

	for rows.Next() {
		var dataset Dataset
		// err := 	rows.Scan(
		// &dataset.accessPolicy, &dataset.createdAt, &dataset.description,
		// &dataset.icon, &dataset.id, &dataset.instance,
		// &dataset.itemCreatedAt, &dataset.itemStatus, &dataset.label,
		// &dataset.label, &dataset.name, &dataset.repositoryURL,
		// &dataset.resourceServer, &dataset.resourceType, &dataset,dataset.resources,
		// &dataset.schema)

		err := 	rows.Scan(&dataset.description, &dataset.label, &dataset.name)

		if err != nil {
			return nil, err
		}

		fmt.Println(dataset)

		datasets = append(datasets, dataset)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return datasets, nil
}

func (g* Dataset) Onboard(app *application.Application) error {
	stmt := `
		INSERT INTO dataset (
			"description",
			"label",
			"name"
		) values (
			$!, $2, $3
		)
	`
	result, err := app.DB.Client.Exec(stmt, g.description, g.label, g.name)

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