package models

import "time"

type Message struct {
	ID            int32      `json:"id,omitempty" db:"id"`
	ChannelID     int32      `json:"channel_id,omitempty" db:"channel_id"`
	Channel       string     `json:"channel" db:"channel_name"`
	GuildID       int32      `json:"guild_id" db:"guild_id"`
	Guild         string     `json:"guild" db:"guild_name"`
	Author        string     `json:"author,omitempty" db:"author"`
	AuthorId      int32      `json:"author_id,omitempty" db:"author_id"`
	Content       string     `json:"content,omitempty" db:"content"`
	AttachmentIDs []int32    `json:"attachment_ids,omitempty"`
	Attachment    []string   `json:"attachments,omitempty"`
	CreatedDate   *time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate   *time.Time `json:"updated_date,omitempty" db:"updated_date"`
}

func (m Message) WebsocketData() {}
