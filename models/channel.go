package models

import "time"

type Channel struct {
	ID           int32     `json:"id,omitempty" db:"id"`
	Name         string    `json:"name,omitempty" db:"name"`
	MessageCount string    `json:"message_count,omitempty" db:"message_count"`
	GuildId      int32     `json:"guild_id,omitempty" db:"guild_id"`
	CreatedDate  time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate  time.Time `json:"updated_date,omitempty" db:"updated_date"`
}
