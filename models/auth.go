package models

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
)

type AuthUser struct {
	ID              int32      `json:"id,omitempty" db:"id"`
	Username        string     `json:"username,omitempty"`
	Password        string     `json:"password,omitempty"`
	ConfirmPassword string     `json:"confirm_password,omitempty"`
	Email           string     `json:"email,omitempty"`
	Avatar          *string    `json:"avatar_url,omitempty" db:"avatar_url"`
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

// This is used while completing the WebAuthn login process.
type BeginLoginResponse struct {
	SessionID string
	*protocol.CredentialAssertion
}
