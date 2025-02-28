package services

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tranquility/config"
	"tranquility/models"

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

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"aud": "https://" + extractPushService(endpoint),
		"exp": time.Now().Add(time.Hour).Unix(),
		"sub": "mailto:your-email@example.com",
	})

	vapidToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return vapidToken, nil
}

func extractPushService(endpoint string) string {
	// Skip the https:// at the beginning and cut of the routes
	return endpoint[8 : strings.Index(endpoint[8:], "/")+8]
}

func (p *PushNotificationService) SendPushNotification(subscription models.PushNotificationInfo, message string) {
	vapidToken, _ := p.generateVAPIDJWT(subscription.Endpoint)

	payload := []byte(message)
	req, _ := http.NewRequest("POST", subscription.Endpoint, bytes.NewBuffer(payload))
	req.Header.Set("TTL", "60")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+vapidToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.ERROR(fmt.Sprintf("push notification failed: %v", err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		p.logger.ERROR(fmt.Sprintf("push notification failed to send with status code: %d", resp.StatusCode))
		return
	}

	p.logger.INFO(fmt.Sprintf("Push notification sent to %d", subscription.UserID))
}
