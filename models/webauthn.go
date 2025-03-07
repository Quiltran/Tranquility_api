package models

import "github.com/go-webauthn/webauthn/webauthn"

type WebAuthnCredential struct {
	Id          int32  `db:"id"`
	Username    string `db:"username"`
	UserHandle  []byte `db:"user_handle"`
	Credentials []webauthn.Credential
}
