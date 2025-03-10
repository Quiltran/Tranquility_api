package models

type Profile struct {
	Username                string  `json:"username,omitempty" db:"username"`
	Email                   string  `json:"email,omitempty" db:"email"`
	NotificationsRegistered bool    `json:"notification_registered,omitempty" db:"notification_registered"`
	AvatarID                *int32  `json:"avatar_id,omitempty" db:"avatar_id"`
	AvatarURL               *string `json:"avatar_url,omitempty" db:"avatar_url"`
}
