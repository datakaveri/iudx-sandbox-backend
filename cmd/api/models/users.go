package models

import (
	"github.com/iudx-sandbox-backend/pkg/application"
)

type User struct {
	UserId    int
	Name      string
	Admin     string
	CreatedAt string
}

func (g *User) Get(app *application.Application, userName string) (User, error) {
	stmt := `
		SELECT "id", "name", "admin", "created"
		FROM users 
		WHERE "name" = $1;
	`

	user := User{}

	row := app.DB.Client.QueryRow(stmt, userName)

	err := row.Scan(&user.UserId, &user.Name, &user.Admin, &user.CreatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}
