package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
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
	// Sessions should not be used twice, so clear the session after being retrieved.
	// If the session was retrieved we can assume it was at least attempted to used.
	delete(w.sessions, key)

	return sessionBytes.Value, nil
}

// # Start should be ran in a goroutine.
//
// The session memory cache should be cleared regularly of expired tokens.
// This clearing should be done in the background so it doesn't block.
// Seeing that sessions are ephemeral due to their expiration time, we do not need to wrap them in a mutex.
func (w *WebAuthnSessions) Start(ctx context.Context, logger Logger) {
	timer := time.NewTicker(time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			var clearedSessionCount int32 = 0
			for key, value := range w.sessions {
				if time.Since(value.InsertedAt) > time.Duration(time.Minute) {
					delete(w.sessions, key)
					clearedSessionCount += 1
				}
			}
			if clearedSessionCount > 0 {
				logger.INFO(fmt.Sprintf("%d WebAuthn sessions have been cleared", clearedSessionCount))
			}
		}
	}
}
