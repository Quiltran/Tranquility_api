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

func (c *Claims) WebAuthnID() []byte {
	authId, err := base64.StdEncoding.DecodeString(c.UserHandle)
	if err != nil {
		log.Println(fmt.Errorf("an error occured while decoding userHandle: %v", err))
		return []byte{}
	}
	return authId
}

func (c *Claims) WebAuthnName() string {
	return c.Username
}

func (c *Claims) WebAuthnDisplayName() string {
	return c.Username
}

func (c *Claims) WebAuthnCredentials() []webauthn.Credential {
	return c.Credentials
}
