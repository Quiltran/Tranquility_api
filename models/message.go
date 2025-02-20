package models

import "time"

type Message struct {
	ID          int32     `json:"id,omitempty" db:"id"`
	ChannelID   int32     `json:"channel_id,omitempty" db:"channel_id"`
	Author      string    `json:"author,omitempty" db:"author"`
	AuthorId    int32     `json:"author_id,omitempty" db:"author_id"`
	Content     string    `json:"content,omitempty" db:"content"`
	CreatedDate time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate time.Time `json:"updated_date,omitempty" db:"updated_date"`
}

func (m Message) WebsocketData() {}
