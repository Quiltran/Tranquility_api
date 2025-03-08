package models

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username    string                `json:"username"`
	ID          int32                 `json:"id" db:"id"`
	UserHandle  string                `json:"userHandle" db:"user_handle"`
	Credentials []webauthn.Credential `json:"-"`
	*jwt.RegisteredClaims
}

// This function is used to implement the webauthn.User interface.
// We cann't change the singature to return an error so the best we can do is log it.
func (c *Claims) WebAuthnID() []byte {
	authId, err := base64.StdEncoding.DecodeString(c.UserHandle)
	if err != nil {
		log.Println(fmt.Errorf("an error occured while decoding userHandle: %v", err))
		return []byte{}
	}
	return authId
}

// This function is used to implement the webauthn.User interface.
func (c *Claims) WebAuthnName() string {
	return c.Username
}

// This function is used to implement the webauthn.User interface.
func (c *Claims) WebAuthnDisplayName() string {
	return c.Username
}

// This function is used to implement the webauthn.User interface.
func (c *Claims) WebAuthnCredentials() []webauthn.Credential {
	return c.Credentials
}
