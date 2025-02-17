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
