package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnCred struct {
	Id          []byte
	Name        string
	DisplayName string
}

func (a *WebAuthnCred) WebAuthnID() []byte {
	return a.Id
}

func (a *WebAuthnCred) WebAuthnName() string {
	return a.Name
}

func (a *WebAuthnCred) WebAuthnDisplayName() string {
	return a.DisplayName
}

func (a *WebAuthnCred) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{}
}
