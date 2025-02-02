package models

import "time"

type CreateMemberRequest struct {
	ID           int32     `json:"id,omitempty"`
	UserId       int       `json:"user_id,omitempty"`
	GuildId      int       `json:"guild_id,omitempty"`
	UserWhoAdded int32     `json:"user_who_added,omitempty"`
	CreatedDate  time.Time `json:"created_date,omitempty"`
	UpdatedDate  time.Time `json:"updated_date,omitempty"`
}
