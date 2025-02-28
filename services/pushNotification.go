package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"tranquility/config"
	"tranquility/models"

	"golang.org/x/crypto/hkdf"

	"github.com/golang-jwt/jwt/v5"
)

type PushNotificationService struct {
	*config.PushNotificationConfig
	logger Logger
}

func (p *PushNotificationService) generateVAPIDJWT(endpoint string) (string, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(p.VapidPrivateKey)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	host, err := extractPushService(endpoint)
	if err != nil {
		return "", fmt.Errorf("error getting host from user's push notification endpoint %s: %v", endpoint, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"aud": host,
		"exp": time.Now().Add(time.Hour).Unix(),
		"sub": "mailto:your-email@example.com",
	})

	vapidToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return vapidToken, nil
}

func extractPushService(endpoint string) (string, error) {
	// Skip the https:// at the beginning and cut of the routes
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

func (p *PushNotificationService) deriveSharedSecret(userP256dh, auth string) ([]byte, error) {
	privateKeyBytes, err := base64.RawURLEncoding.DecodeString(p.VapidPrivateKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := ecdh.P256().NewPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	userPubKeyBytes, err := base64.RawURLEncoding.DecodeString(userP256dh)
	if err != nil {
		return nil, err
	}
	if len(userPubKeyBytes) != 65 || userPubKeyBytes[0] != 0x04 {
		return nil, errors.New("invalid user public key format")
	}

	userPubKey, err := privateKey.PublicKey().Curve().NewPublicKey(userPubKeyBytes[1:])
	if err != nil {
		return nil, err
	}

	sharedSecret, err := privateKey.ECDH(userPubKey)
	if err != nil {
		return nil, err
	}

	authSecret, err := base64.RawURLEncoding.DecodeString(auth)
	if err != nil {
		return nil, err
	}

	prk := hkdfExtract(sharedSecret, authSecret)
	return prk, nil
}

func hkdfExtract(sharedSecret, authSecret []byte) []byte {
	salt := make([]byte, 32)

	hkdf := hkdf.New(sha256.New, sharedSecret, salt, authSecret)
	prk := make([]byte, 32)
	_, _ = hkdf.Read(prk)

	return prk
}

func deriveEncryptionKey(sharedSecret, salt []byte, info string, length int) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, sharedSecret, salt, []byte(info))
	key := make([]byte, length)
	_, err := hkdf.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (p *PushNotificationService) encryptPayload(subscription *models.PushNotificationInfo, message string) ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	sharedSecret, err := p.deriveSharedSecret(subscription.P256dh, subscription.Auth)
	if err != nil {
		return nil, err
	}

	encryptionKey, err := deriveEncryptionKey(sharedSecret, salt, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}

	nonce, err := deriveEncryptionKey(sharedSecret, salt, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encrypted := aesgcm.Seal(nil, nonce, []byte(message), nil)

	finalPayload := append(salt, nonce...)
	finalPayload = append(finalPayload, encrypted...)

	return finalPayload, nil
}

func (p *PushNotificationService) SendPushNotification(subscription models.PushNotificationInfo, message string) {
	vapidToken, err := p.generateVAPIDJWT(subscription.Endpoint)
	if err != nil {
		p.logger.ERROR(fmt.Sprintf("push notification failed generate jwt: %v", err))
		return
	}

	payload, err := p.encryptPayload(&subscription, message)
	if err != nil {
		p.logger.ERROR(fmt.Sprintf("push notification failed to encrypt: %v", err))
		return
	}
	req, _ := http.NewRequest("POST", subscription.Endpoint, bytes.NewBuffer(payload))
	req.Header.Set("TTL", "60")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+vapidToken)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(payload)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.ERROR(fmt.Sprintf("push notification failed: %v", err))
		return
	}
	if resp.StatusCode == http.StatusGone {
		p.logger.ERROR(fmt.Sprintf("push notification failed to send with status code: %d", resp.StatusCode))
		return
	} else if resp.StatusCode != http.StatusCreated {
		p.logger.ERROR(fmt.Sprintf("push notification failed to send with status code: %d", resp.StatusCode))
		return
	}

	p.logger.INFO(fmt.Sprintf("Push notification sent to %d", subscription.UserID))
}
