package data

import (
	"context"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type authRepo struct {
	db *sqlx.DB
}

func (a *authRepo) Login(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		"SELECT id, username, password, refresh_token, websocket_token FROM auth WHERE username = $1",
		// "INSERT INTO auth (username, password, email) VALUES (:username, :last_name, :email) RETURNING id, username, email, refresh_token;",
		&user.Username,
	).StructScan(&output)

	return &output, err
}

func (a *authRepo) Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		"INSERT INTO auth (username, password, email) VALUES ($1, $2, $3) RETURNING id, username, email, refresh_token, created_date;",
		user.Username,
		user.Password,
		user.Email,
	).StructScan(&output)

	return &output, err
}
