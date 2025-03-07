package services

import (
	"crypto/rand"
	"errors"
	"time"
	"tranquility/config"

	"github.com/go-webauthn/webauthn/webauthn"
)

var (
	ErrSessionNotFound = errors.New("the WebAuthn session request was not found")
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

func GenerateWebAuthnID() ([]byte, error) {
	id := make([]byte, 64)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}

	return id, nil
}

type webAuthnSessionData struct {
	Value      []byte
	InsertedAt time.Time
}

type WebAuthnSessions struct {
	sessions map[string]webAuthnSessionData
}

func NewWebAuthnSessions() *WebAuthnSessions {
	return &WebAuthnSessions{
		sessions: make(map[string]webAuthnSessionData),
	}
}

func (w *WebAuthnSessions) AddSession(key string, sessionBytes []byte) {
	w.sessions[key] = webAuthnSessionData{sessionBytes, time.Now()}
}

func (w *WebAuthnSessions) GetSession(key string) ([]byte, error) {
	sessionBytes, ok := w.sessions[key]
	if !ok {
		return nil, ErrSessionNotFound
	}
	delete(w.sessions, key)

	return sessionBytes.Value, nil
}

func (w *WebAuthnSessions) ClearExpiredSessions() int32 {
	var output int32 = 0
	for key, value := range w.sessions {
		if time.Since(value.InsertedAt) > time.Duration(time.Minute) {
			delete(w.sessions, key)
			output += 1
		}
	}

	return output
}
