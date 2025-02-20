package models

import "time"

type Message struct {
	ID          int32     `json:"id,omitempty"`
	ChannelID   int32     `json:"channel_id,omitempty"`
	Author      string    `json:"author,omitempty"`
	AuthorId    int32     `json:"author_id,omitempty"`
	Content     string    `json:"content,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
	UpdatedDate time.Time `json:"updated_date,omitempty"`
}

func (m Message) WebsocketData() {}
