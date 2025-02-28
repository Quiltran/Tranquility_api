package data

import (
	"context"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type messageRepo struct {
	db *sqlx.DB
}

func (m *messageRepo) GetChannelMessages(ctx context.Context, userId, guildId, channelId, pageNumber int32) ([]models.Message, error) {
	var output []models.Message

	pageSize := int64(20)
	offset := int64(pageNumber) * pageSize

	rows, err := m.db.QueryContext(
		ctx,
		`SELECT
			m.id,
			m.channel_id,
			a.username as author,
			m.author_id,
			m.content,
			m.created_date,
			m.updated_date
		FROM message m
		JOIN auth a on a.id = m.author_id
		JOIN channel c ON m.channel_id = c.id
		JOIN guild g ON c.guild_id = g.id
		JOIN member mem ON mem.guild_id = g.id
		WHERE   g.id = $1
			AND c.id = $2
			AND mem.user_id = $3
		ORDER BY m.id ASC
		OFFSET $4
		LIMIT $5;`,
		guildId,
		channelId,
		userId,
		offset,
		pageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message models.Message
		if err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.Author,
			&message.AuthorId,
			&message.Content,
			&message.CreatedDate,
			&message.UpdatedDate,
		); err != nil {
			return nil, err
		}
		output = append(output, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, nil
}

func (m *messageRepo) CreateMessage(ctx context.Context, message *models.Message, userId int32) (*models.Message, error) {
	var output models.Message

	err := m.db.QueryRowxContext(
		ctx,
		`WITH im AS (
        INSERT INTO message (author_id, channel_id, content)
        SELECT $1, $2, $3
        WHERE EXISTS (
            SELECT 1 FROM channel c
            join member m on c.guild_id = m.guild_id
            where m.user_id = $1 and c.id = $2
        )
        RETURNING id, channel_id, author_id, content, created_date, updated_date
        )
        SELECT
            im.id,
            im.channel_id,
            a.username as author,
            im.author_id,
            im.content,
            im.created_date,
            im.updated_date
        FROM im
        JOIN auth a ON im.author_id = a.id`,
		userId,
		message.ChannelID,
		message.Content,
	).StructScan(&output)

	if err != nil {
		return nil, err
	}

	return &output, nil
}
