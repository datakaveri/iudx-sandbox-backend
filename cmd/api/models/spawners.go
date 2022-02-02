package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
)

type Spawner struct {
	Id       int
	UserId   int
	ServerId int
}

func (g *Spawner) GetSpawnerIdBasedOnBaseUrl(app *application.Application, baseUrl string, userId int) (Spawner, error) {
	stmt := `
		SELECT spawners."id", spawners."user_id", spawners."server_id"
		FROM spawners
		LEFT JOIN servers
		ON servers."id" = spawners."server_id"
		WHERE servers."base_url" = $1 AND spawners."user_id" = $2
	`

	spawner := Spawner{}

	row := app.DB.Client.QueryRow(stmt, baseUrl, userId)

	err := row.Scan(&spawner.Id, &spawner.UserId, &spawner.ServerId)
	if err != nil {
		return spawner, err
	}

	return spawner, nil
}
