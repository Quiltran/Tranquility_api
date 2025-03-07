package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tranquility/models"

	"github.com/go-webauthn/webauthn/webauthn"
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

func (a *authRepo) saveWebAuthnRegistrationSession(ctx context.Context, userId int32, sessionData []byte) error {
	tx, err := a.db.Beginx()
	if err != nil {
		return fmt.Errorf("an error occurred while beginning transaction to save webauthn registration session: %v", err)
	}
	defer tx.Rollback()

	affected, err := tx.ExecContext(
		ctx,
		`INSERT INTO webauthn_cache (key,value) VALUES ($1, $2);`,
		&userId,
		&sessionData,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while saving webauthn registration session: %v", err)
	}

	rows, err := affected.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while getting number of rows affected by saving webauthn registration session: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("more than one row was affected when saving webauthn registration session: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occured while commiting saving webauthn registration session: %v", err)
	}
	return nil
}

func (a *authRepo) getWebAuthnRegistrationSession(ctx context.Context, userId int32) ([]byte, error) {
	var session []byte

	err := a.db.QueryRowxContext(
		ctx,
		`SELECT value FROM webauthn_cache WHERE key = $1`,
		&userId,
	).Scan(&session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (a *authRepo) saveWebAuthnCredential(ctx context.Context, credentials *webauthn.Credential, userId int32) error {
	tx, err := a.db.Beginx()
	if err != nil {
		return fmt.Errorf("an error occurred while beginning a transaction to save webauthn credentials: %v", err)
	}
	defer tx.Rollback()

	affected, err := tx.ExecContext(
		ctx,
		`INSERT INTO webauthn_credentials (
			user_id,
			credential_id,
			public_key,
			signature_count
		) VALUES ($1, $2, $3, $4)`,
		&userId,
		&credentials.ID,
		&credentials.PublicKey,
		&credentials.Authenticator.SignCount,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while saving auth credentials after completing webauthn registration: %v", err)
	}

	rows, err := affected.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while getting number of rows affected while saving webauthn credentials: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("more than one record was affected while saving webauthn credentials: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occured while commiting saving webauthn credentials: %v", err)
	}
	return nil
}
