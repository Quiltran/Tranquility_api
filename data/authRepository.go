package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		&user.Username,
	).StructScan(&output)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}

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

func (a *authRepo) RefreshToken(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		"UPDATE auth SET refresh_token = md5(random()::text), websocket_token = md5(random()::text), updated_date = NOW() AT TIME ZONE 'utc' WHERE id = $1 AND refresh_token = $2 RETURNING id, username, email, refresh_token, websocket_token, updated_date;",
		user.ID,
		user.RefreshToken,
	).StructScan(&output)

	return &output, err
}

func (a *authRepo) WebsocketLogin(ctx context.Context, userId int32, websocketToken string) (*models.AuthUser, error) {
	var output models.AuthUser

	err := a.db.QueryRowxContext(
		ctx,
		`UPDATE auth SET websocket_token = md5(random()::text) WHERE id = $1 AND websocket_token = $2
		RETURNING id, username, refresh_token, websocket_token;`,
		userId,
		websocketToken,
	).StructScan(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (a *authRepo) GetUserProfile(ctx context.Context, userId int32) (*models.Profile, error) {
	var output models.Profile

	err := a.db.QueryRowxContext(
		ctx,
		`SELECT
			a.username,
			a.email,
			EXISTS(SELECT 1 FROM notification WHERE user_id = a.id) AS notification_registered
		FROM auth a WHERE a.id = $1`,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, fmt.Errorf("An error occurred while collecting profile information for %d: %v", userId, err)
	}

	return &output, nil
}
