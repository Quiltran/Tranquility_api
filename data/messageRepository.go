package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type messageRepo struct {
	db *sqlx.DB
}

func (m *messageRepo) GetChannelMessages(ctx context.Context, userId, guildId, channelId, pageNumber int32) ([]models.Message, error) {
	var output []models.Message = make([]models.Message, 0)

	pageSize := int64(50)
	offset := int64(pageNumber) * pageSize

	rows, err := m.db.QueryContext(
		ctx,
		`SELECT
			m.id,
			m.channel_id,
			m.author,
			m.author_id,
			m.content,
			m.created_date,
			m.updated_date
		FROM (
			SELECT
				m.id,
				m.channel_id,
				a.username AS author,
				m.author_id,
				m.content,
				m.created_date,
				m.updated_date
			FROM message m
			JOIN auth a ON a.id = m.author_id
			JOIN channel c ON m.channel_id = c.id
			JOIN guild g ON c.guild_id = g.id
			JOIN member mem ON mem.guild_id = g.id
			WHERE g.id = $1
			AND c.id = $2
			AND mem.user_id = $3
			ORDER BY m.id DESC
			OFFSET $4
			LIMIT $5
		) AS m
		ORDER BY created_date ASC;`,
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

func (m *messageRepo) CreateMessage(ctx context.Context, message *models.Message, userId int32) (*sqlx.Tx, *models.Message, error) {
	var output models.Message

	tx, err := m.db.Beginx()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create transaction while creating message")
	}

	err = tx.QueryRowxContext(
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
			c.name as channel_name,
			g.id as guild_id,
			g.name as guild_name,
            a.username as author,
            im.author_id,
            im.content,
            im.created_date,
            im.updated_date
        FROM im
        JOIN auth a ON im.author_id = a.id
		JOIN channel c on c.id = im.channel_id
		JOIN guild g on c.guild_id = g.id`,
		userId,
		message.ChannelID,
		message.Content,
	).StructScan(&output)

	if err != nil {
		return nil, nil, err
	}

	return tx, &output, nil
}

func (m *messageRepo) CreateAttachmentMapping(ctx context.Context, tx *sqlx.Tx, messageId, attachmentId int32) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO attachment_mapping (post_id, attachment_id) VALUES ($1, $2)`,
		&messageId,
		&attachmentId,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while inserting into attachment mapping: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while getting the number of rows affected while inserting into attachment mapping: %s", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("more than one attachment mapping was created while inserting attachment for %d", messageId)
	}

	return nil
}

func (m *messageRepo) GetMessageAttachment(ctx context.Context, messageID int32) ([]models.Attachment, error) {
	var attachments []models.Attachment
	rows, err := m.db.QueryxContext(
		ctx,
		`SELECT id, file_name, file_path
		FROM attachment a
		JOIN attachment_mapping am on am.attachment_id = a.id
		WHERE am.post_id = $1`,
		&messageID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("an error occurred while selecting message attachment information: %v", err)
	}

	for rows.Next() {
		var attachment models.Attachment
		if err := rows.Scan(
			&attachment.ID,
			&attachment.FileName,
			&attachment.FilePath,
		); err != nil {
			return nil, fmt.Errorf("an error occurred while scanning attachment information: %v", err)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}
