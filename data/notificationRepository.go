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

func (n *notificationRepo) GetNotificationRecipients(ctx context.Context, userId, channelId int32) ([]models.PushNotificationInfo, error) {
	var output []models.PushNotificationInfo

	rows, err := n.db.QueryxContext(
		ctx,
		`SELECT n.user_id, n.endpoint, n.p256dh, n.auth
		FROM notification n
		JOIN member m on m.user_id = n.user_id
		JOIN channel c on c.guild_id = m.guild_id
		WHERE c.id = $1 and n.user_id != $2`,
		channelId,
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var info models.PushNotificationInfo
		if err := rows.StructScan(&info); err != nil {
			return nil, err
		}

		output = append(output, info)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, nil
}

func (n *notificationRepo) DeleteUserPushInformation(ctx context.Context, userId int32) error {
	tx, err := n.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction for deleting push notifications: %v", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM notification WHERE user_id = $1`,
		userId,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while executing delete push notification: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while collecting number of push notifications deleted: %v", err)
	}
	if affected != 1 {
		return fmt.Errorf("an invalid number of push notification rows were deleted")
	}

	tx.Commit()
	return nil
}
