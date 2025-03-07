package services

import (
	"crypto/rand"
	"tranquility/config"
	"tranquility/models"

	"github.com/go-webauthn/webauthn/webauthn"
)

func NewWebauthn(config *config.WebAuthnConfig) (*webauthn.WebAuthn, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     config.RPOrigins,
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return webAuthn, nil
}

func generateWebAuthnID() ([]byte, error) {
	id := make([]byte, 64)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func NewWebAuthnCred(name, displayName string) (*models.WebAuthnCred, error) {
	id, err := generateWebAuthnID()
	if err != nil {
		return nil, err
	}

	return &models.WebAuthnCred{
		id,
		name,
		displayName,
	}, nil
}
