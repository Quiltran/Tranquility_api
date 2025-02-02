package models

import "time"

type CreateGuildRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateGuildResponse struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	OwnerID     int       `json:"ownder_id,omitempty"`
	Channels    []Channel `json:"channels,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
	UpdatedDate time.Time `json:"updated_date,omitempty"`
}
