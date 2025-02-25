package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type memberRepo struct {
	db *sqlx.DB
}

func (m *memberRepo) addGuildMember(ctx context.Context, guildId, userId int32, tx *sqlx.Tx) error {
	query := `INSERT INTO member (user_id, guild_id, user_who_added) VALUES ($1, $2, $3) RETURNING id, user_id, guild_id, user_who_added, created_date, updated_date;`
	var rows sql.Result
	var err error

	if tx != nil {
		rows, err = tx.ExecContext(ctx, query, userId, guildId, userId)
	} else {
		rows, err = m.db.ExecContext(ctx, query, userId, guildId, userId)
	}

	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("an invalid number rows were affected while inserting member")
	}

	return nil
}

func (m *memberRepo) CreateMember(ctx context.Context, member *models.Member) (*models.Member, error) {
	var output models.Member
	err := m.db.QueryRowxContext(
		ctx,
		`INSERT INTO member (user_id, guild_id, user_who_added)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, guild_id, user_who_added, created_date, updated_date;`,
		&member.UserId,
		&member.GuildId,
		&member.UserWhoAdded,
	).StructScan(&output)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, ErrDuplicateMember
			}
		}
		return nil, err
	}

	return &output, nil
}

func (m *memberRepo) GetGuildMembers(ctx context.Context, guildId, userId int32) ([]models.AuthUser, error) {
	output := []models.AuthUser{}

	rows, err := m.db.QueryContext(
		ctx,
		`SELECT a.id, a.Username
		FROM member m JOIN auth a ON a.id = m.user_id
		WHERE m.guild_id = $1
		AND EXISTS(
			SELECT 1
			FROM member requester
			WHERE requester.user_id = $2 AND requester.guild_id = $1
		);`,
		&guildId,
		&userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.AuthUser
		if err := rows.Scan(&member.ID, &member.Username); err != nil {
			return nil, err
		}
		output = append(output, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, sql.ErrNoRows
	}

	return output, nil
}

func (m *memberRepo) GetChannelMembers(ctx context.Context, channelId int32) (map[int32]bool, error) {
	output := make(map[int32]bool)

	rows, err := m.db.QueryContext(
		ctx,
		`SELECT m.user_id FROM member m
		 JOIN channel c ON c.guild_id = m.guild_id
		 WHERE c.id = $1`,
		&channelId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userId int32
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		output[userId] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, nil
}
