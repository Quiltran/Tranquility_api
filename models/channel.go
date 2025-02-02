package models

import "time"

type Channel struct {
	ID           int32     `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	MessageCount string    `json:"message_count,omitempty"`
	GuildId      int32     `json:"guild_id,omitempty"`
	CreatedDate  time.Time `json:"created_date,omitempty"`
	UpdatedDate  time.Time `json:"updated_date,omitempty"`
}
