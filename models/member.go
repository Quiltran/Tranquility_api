package models

import "time"

type Member struct {
	ID           int32      `json:"id,omitempty" db:"id"`
	UserId       int        `json:"user_id,omitempty" db:"user_id"`
	Username     string     `json:"username"`
	GuildId      int        `json:"guild_id,omitempty" db:"guild_id"`
	UserWhoAdded int32      `json:"user_who_added,omitempty" db:"user_who_added"`
	CreatedDate  *time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate  *time.Time `json:"updated_date,omitempty" db:"updated_date"`
}
