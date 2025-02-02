package models

import "time"

type Intent struct {
	ID          int32     `json:"id,omitempty"`
	RoleID      int32     `json:"role_id,omitempty"`
	Value       int32     `json:"value,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
	UpdatedDate time.Time `json:"updated_date,omitempty"`
}

type Role struct {
	ID          int32     `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	GuildId     int32     `json:"guild_id,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
	UpdatedDate time.Time `json:"updated_date,omitempty"`
}
