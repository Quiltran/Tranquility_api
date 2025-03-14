package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ConnectionString string
	UploadPath       string
	AllowedOrigins   []string
	TurnstileSecret  string
	*JWTConfig
	*PushNotificationConfig
	*WebAuthnConfig
}

type JWTConfig struct {
	JWEPrivateKey *rsa.PrivateKey
	Lifetime      time.Duration
	Issuer        string
	Audience      []string
	Key           string
}

type WebAuthnConfig struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

type PushNotificationConfig struct {
	VapidPrivateKey string
	VapidPublicKey  string
	Sub             string
}

func NewConfig() (*Config, error) {
	connectionString := os.Getenv("CONNECTION_STRING")
	if connectionString == "" {
		return nil, errors.New("CONNECTION_STRING was not set")
	}

	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		return nil, errors.New("UPLOAD_PATH was not set")
	}

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(origins) == 0 || origins[0] == "" {
		return nil, errors.New("ALLOWED_ORIGINS was not set")
	}

	turnstileSecret := os.Getenv("TURNSTILE_SECRET")
	if turnstileSecret == "" {
		return nil, errors.New("TURNSTILE_SECRET was not set")
	}

	jwtConfig, err := loadJWTConfig()
	if err != nil {
		return nil, err
	}
	pushNotificationConfig, err := loadPushNotificationConfig()
	if err != nil {
		return nil, err
	}

	webAuthnConfig, err := loadWebAuthnConfig()
	if err != nil {
		return nil, err
	}
	return &Config{
		ConnectionString:       connectionString,
		UploadPath:             uploadPath,
		AllowedOrigins:         origins,
		TurnstileSecret:        turnstileSecret,
		JWTConfig:              jwtConfig,
		PushNotificationConfig: pushNotificationConfig,
		WebAuthnConfig:         webAuthnConfig,
	}, nil
}

func loadJWTConfig() (*JWTConfig, error) {
	jwePem, err := os.ReadFile(os.Getenv("JWT_PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, fmt.Errorf("an error occurred while reading JWE private key: %v", err)
	}
	block, _ := pem.Decode(jwePem)
	if block == nil {
		return nil, fmt.Errorf("an error occurred while decoding JWE private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing JWE private key: %v", err)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("an error occurred while converting JWE private key to RSA: %v", err)
	}

	lifetimeSetting := os.Getenv("JWT_LIFETIME")
	var lifetime time.Duration
	if lifetimeSetting != "" {
		l, err := strconv.ParseInt(lifetimeSetting, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("an error occurred while loading jwt lifetime: %v", err)
		}
		lifetime = time.Duration(time.Duration(l) * time.Minute)
	}
	if lifetime == 0 {
		lifetime = time.Duration(2 * time.Minute)
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		return nil, fmt.Errorf("JWT_ISSUER was not set")
	}

	audienceSetting := os.Getenv("JWT_AUDIENCE")
	if audienceSetting == "" {
		return nil, fmt.Errorf("JWT_AUDIENCE was not set")
	}
	audience := strings.Split(audienceSetting, ",")
	slices.Sort(audience)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET was not set")
	}

	jwtConfig := &JWTConfig{
		Lifetime:      time.Duration(2 * time.Minute),
		Issuer:        issuer,
		Audience:      audience,
		Key:           jwtSecret,
		JWEPrivateKey: rsaKey,
	}

	return jwtConfig, nil
}

func loadPushNotificationConfig() (*PushNotificationConfig, error) {
	privateVapidKey := os.Getenv("VAPID_PRIVATE")
	if privateVapidKey == "" {
		return nil, fmt.Errorf("VAPID_PRIVATE was not set")
	}

	publicVapidKey := os.Getenv("VAPID_PUBLIC")
	if publicVapidKey == "" {
		return nil, fmt.Errorf("VAPID_PUBLIC was not set")
	}

	pushNotificationSub := os.Getenv("PUSH_SUB")
	if pushNotificationSub == "" {
		return nil, fmt.Errorf("PUSH_SUB was not set")
	}

	return &PushNotificationConfig{
		VapidPrivateKey: privateVapidKey,
		VapidPublicKey:  publicVapidKey,
		Sub:             pushNotificationSub,
	}, nil
}

func loadWebAuthnConfig() (*WebAuthnConfig, error) {
	displayName := os.Getenv("RP_DISPLAY_NAME")
	if displayName == "" {
		return nil, fmt.Errorf("RP_DISPLAY_NAME was not set")
	}

	rpId := os.Getenv("RPID")
	if rpId == "" {
		return nil, fmt.Errorf("RPID was not set")
	}

	rpOrigins := strings.Split(os.Getenv("RP_ORIGINS"), ",")
	for i := range rpOrigins {
		if rpOrigins[i] == "" {
			return nil, fmt.Errorf("invalid RP_ORIGINS were dectected: %+v", rpOrigins)
		}
	}
	return &WebAuthnConfig{
		RPDisplayName: displayName,
		RPID:          rpId,
		RPOrigins:     rpOrigins,
	}, nil
}
