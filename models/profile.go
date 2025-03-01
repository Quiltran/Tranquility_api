package models

type Profile struct {
	Username                string `json:"username,omitempty" db:"username"`
	Email                   string `json:"email,omitempty" db:"email"`
	NotificationsRegistered bool   `json:"notification_registered,omitempty" db:"notification_registered"`
}
