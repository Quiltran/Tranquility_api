package models

import (
	"time"
)

type AuthUser struct {
	ID              int32      `json:"id,omitempty" db:"id"`
	Username        string     `json:"username,omitempty"`
	Password        string     `json:"password,omitempty"`
	ConfirmPassword string     `json:"confirm_password,omitempty"`
	Email           string     `json:"email,omitempty"`
	Token           string     `json:"token,omitempty"`
	RefreshToken    string     `json:"refresh_token,omitempty" db:"refresh_token"`
	WebsocketToken  string     `json:"websocket_token,omitempty" db:"websocket_token"`
	Turnstile       string     `json:"turnstile,omitempty"`
	UserHandle      []byte     `json:"userHandle,omitempty" db:"user_handle"`
	CreatedDate     *time.Time `json:"created_date,omitempty" db:"created_date"`
	UpdatedDate     *time.Time `json:"updated_date,omitempty" db:"updated_date"`
}

func (a *AuthUser) ClearAuth() {
	a.ID = 0
	a.Password = ""
	a.ConfirmPassword = ""
	a.UserHandle = nil
}
