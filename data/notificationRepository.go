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

func (n *notificationRepo) SaveUserPushInformation(ctx context.Context, registration models.PushNotificationRegistration, userId int32) error {
	tx, err := n.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO notification (user_id, endpoint, pd256dh, auth) VALUES ($1, $2, $3, $4)`,
		&userId,
		&registration.Endpoint,
		&registration.Keys.P256dh,
		&registration.Keys.Auth,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("an error occurred while rolling back %d register for push notifications: %v", userId, err)
		}
		return err
	}

	if affected != 1 {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("an error occurred while rolling back %d register for push notifications: %v", userId, err)
		}
		return fmt.Errorf("an invalid number of rows were affected while adding %d to push notifications", userId)
	}
	return nil
}

func (n *notificationRepo) GetUserPushNotificationInfo(ctx context.Context, userId int32) (*models.PushNotificationInfo, error) {
	var output models.PushNotificationInfo
	err := n.db.QueryRowxContext(
		ctx,
		`SELECT user_id, endpoint, pd256dh, auth
		FROM notification
		WHERE user_id = $1`,
		userId,
	).StructScan(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
