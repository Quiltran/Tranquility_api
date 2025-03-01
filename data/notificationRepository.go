package data

import (
	"context"
	"fmt"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type notificationRepo struct {
	db *sqlx.DB
}

func (n *notificationRepo) SaveUserPushInformation(ctx context.Context, registration *models.PushNotificationRegistration, userId int32) (*sqlx.Tx, error) {
	tx, err := n.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO notification (user_id, endpoint, p256dh, auth) VALUES ($1, $2, $3, $4)`,
		&userId,
		&registration.Endpoint,
		&registration.Keys.P256dh,
		&registration.Keys.Auth,
	)
	if err != nil {
		if txerr := tx.Rollback(); txerr != nil {
			return nil, fmt.Errorf("an error occurred while rolling back %d register for push notifications: %v", userId, txerr)
		}
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		if txerr := tx.Rollback(); txerr != nil {
			return nil, fmt.Errorf("an error occurred while rolling back %d register for push notifications: %v", userId, txerr)
		}
		return nil, err
	}

	if affected != 1 {
		if err := tx.Rollback(); err != nil {
			return nil, fmt.Errorf("an error occurred while rolling back %d register for push notifications: %v", userId, err)
		}
		return nil, fmt.Errorf("an invalid number of rows were affected while adding %d to push notifications", userId)
	}
	return tx, nil
}

func (n *notificationRepo) GetUserPushNotificationInfo(ctx context.Context, userId int32) (*models.PushNotificationInfo, error) {
	var output models.PushNotificationInfo
	err := n.db.QueryRowxContext(
		ctx,
		`SELECT user_id, endpoint, p256dh, auth
		FROM notification
		WHERE user_id = $1`,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
