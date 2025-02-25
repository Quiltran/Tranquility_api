package models

import "time"

type Guild struct {
	ID          int32      `json:"id,omitempty" db:"id"`
	Name        string     `json:"name,omitempty" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	OwnerId     int32      `json:"owner_id,omitempty" db:"owner_id"`
	Channels    []Channel  `json:"channels,omitempty" db:"channels"`
	Members     []AuthUser `json:"members,omitempty" db:"members"`
	CreatedDate *time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate *time.Time `json:"updated_date,omitempty" db:"updated_date"`
}
