package data

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"tranquility/models"
	"tranquility/services"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrDuplicateProfileAttachment = errors.New("the user is already using this file")
)

type authRepo struct {
	db *sqlx.DB
}

func (a *authRepo) Login(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		"SELECT id, username, password, refresh_token, websocket_token, user_handle FROM auth WHERE username = $1",
		&user.Username,
	).StructScan(&output)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}

	return &output, err
}

// # THIS FUNCTION SHOULD NOT BE CALLED IN NORMAL CIRCUMSTANCES
//
// UpdateLoginUserHandle is used to give a user a handle that is used with WebAuthn.
// If the user's user handle is changed, their credentials stored on their device will no longer work.
func (a *authRepo) UpdateLoginUserHandle(ctx context.Context, userId int32) error {
	userHandle, err := services.GenerateWebAuthnID()
	if err != nil {
		return fmt.Errorf("an error occurred while updating user handle: %v", err)
	}

	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("an error occurred while beginning transaction to update user handle: %v", err)
	}
	defer tx.Rollback()

	rows, err := tx.ExecContext(
		ctx,
		`UPDATE auth SET user_handle = $1 WHERE id = $2`,
		userHandle,
		userId,
	)
	if err != nil {
		return fmt.Errorf("an error occured while updating user_handle for %d: %v", userId, err)
	}

	if affected, err := rows.RowsAffected(); err != nil {
		return fmt.Errorf("an error occurred while getting the number of rows affected while updating user_handle for %d: %v", userId, err)
	} else if affected != 1 {
		return fmt.Errorf("more than 1 row was affected while updating user_handle for %d: %v", userId, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit transaction while updating user_handle: %v", err)
	}
	return nil
}

func (a *authRepo) Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	userHandler, err := services.GenerateWebAuthnID()
	if err != nil {
		return nil, fmt.Errorf("an error occurred while generating webauthn user_handle: %v", err)
	}
	err = a.db.QueryRowxContext(
		ctx,
		"INSERT INTO auth (username, password, email, user_handle) VALUES ($1, $2, $3, $4) RETURNING id, username, email, refresh_token, created_date, user_handle;",
		user.Username,
		user.Password,
		user.Email,
		userHandler,
	).StructScan(&output)

	return &output, err
}

func (a *authRepo) RefreshToken(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	var output models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		"UPDATE auth SET refresh_token = md5(random()::text), websocket_token = md5(random()::text), updated_date = NOW() AT TIME ZONE 'utc' WHERE id = $1 AND refresh_token = $2 RETURNING id, username, email, refresh_token, websocket_token, updated_date, user_handle;",
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
			at.file_name as avatar_url,
			EXISTS(SELECT 1 FROM notification WHERE user_id = a.id) AS notification_registered
		FROM auth a
		LEFT JOIN profile_mapping pm on pm.user_id = a.id
		LEFT JOIN attachment at on pm.attachment_id = at.id
		WHERE a.id = $1`,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while collecting profile information for %d: %v", userId, err)
	}

	return &output, nil
}

func (a *authRepo) UpdateUserProfile(ctx context.Context, tx *sqlx.Tx, profile *models.Profile, userId int32) error {
	rows, err := tx.ExecContext(
		ctx,
		`UPDATE auth
		SET username = $1, email = $2
		WHERE id = $3`,
		&profile.Username,
		&profile.Email,
		&userId,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while updating the user profile %d: %v", userId, err)
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while getting the number of affected profiles updated for %d: %v", userId, err)
	} else if affected != 1 {
		return fmt.Errorf("an invalid number of profiles were updated while updating %d: %v", userId, err)
	}

	return nil
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
			signature_count,
			backup_eligible,
			backup_state
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		&userId,
		&credentials.ID,
		&credentials.PublicKey,
		&credentials.Authenticator.SignCount,
		&credentials.Flags.BackupEligible,
		&credentials.Flags.BackupState,
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

func (a *authRepo) getWebAuthnCredential(ctx context.Context, rawId, userHandle []byte) (*models.AuthUser, *models.Claims, error) {
	var userCredentials models.AuthUser
	err := a.db.QueryRowxContext(
		ctx,
		`SELECT a.id, a.username, a.refresh_token, a.websocket_token, a.user_handle
		FROM auth a
		JOIN webauthn_credentials wc on wc.user_id = a.id
		WHERE wc.credential_id = $1 and a.user_handle = $2`,
		rawId,
		userHandle,
	).StructScan(&userCredentials)
	if err != nil {
		return nil, nil, err
	}

	var credential webauthn.Credential
	err = a.db.QueryRowxContext(
		ctx,
		`SELECT credential_id, public_key, signature_count, backup_eligible, backup_state from webauthn_credentials WHERE credential_id = $1 and user_id = $2`,
		rawId,
		userCredentials.ID,
	).Scan(
		&credential.ID,
		&credential.PublicKey,
		&credential.Authenticator.SignCount,
		&credential.Flags.BackupEligible,
		&credential.Flags.BackupState,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("an error occurred while collecting the credential for webauthn login for %s: %v", userCredentials.Username, err)
	}

	return &userCredentials,
		&models.Claims{
			ID:          userCredentials.ID,
			Username:    userCredentials.Username,
			UserHandle:  base64.StdEncoding.EncodeToString(userCredentials.UserHandle),
			Credentials: []webauthn.Credential{credential},
		},
		nil
}

func (a *authRepo) CreateProfileMapping(ctx context.Context, tx *sqlx.Tx, profileId, attachmentId int32) (*string, error) {
	var replacedFileName *string

	var oldAttachmentId int32
	err := tx.QueryRowxContext(
		ctx,
		`SELECT a.file_name, pm.attachment_id
		FROM profile_mapping as pm
		JOIN attachment a on a.id = pm.attachment_id
		WHERE user_id = $1`,
		profileId,
	).Scan(&replacedFileName, &oldAttachmentId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if oldAttachmentId == attachmentId {
		return nil, ErrDuplicateProfileAttachment
	}

	rows, err := tx.ExecContext(
		ctx,
		`WITH updated_row AS (
			INSERT INTO profile_mapping (user_id, attachment_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE
				SET attachment_id = $2
			WHERE profile_mapping.attachment_id != $2
			RETURNING attachment_id
		)
		SELECT a.file_name
		FROM attachment a
		JOIN updated_row ur on a.id = ur.attachment_id`,
		profileId,
		attachmentId,
	)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while upserting profile mapping for %d: %v", profileId, err)
	}

	if affected, err := rows.RowsAffected(); err != nil {
		return nil, fmt.Errorf("an error occurred while getting the number of rows affected by %d profile update: %v", profileId, err)
	} else if affected != 1 {
		return nil, fmt.Errorf("more than one profile was affected by %d profile update", profileId)
	}

	return replacedFileName, nil
}
