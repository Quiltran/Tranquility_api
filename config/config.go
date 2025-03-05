package config

import (
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
	Lifetime time.Duration
	Issuer   string
	Audience []string
	Key      string
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

	return &Config{
		ConnectionString:       connectionString,
		UploadPath:             uploadPath,
		AllowedOrigins:         origins,
		TurnstileSecret:        turnstileSecret,
		JWTConfig:              loadJWTConfig(),
		PushNotificationConfig: loadPushNotificationConfig(),
		WebAuthnConfig:         loadWebAuthnConfig(),
	}, nil
}

func loadJWTConfig() *JWTConfig {
	jwtConfig := &JWTConfig{
		Lifetime: time.Duration(2 * time.Minute),
		Issuer:   "api.quiltran.com",
		Audience: []string{"quiltran.com"},
		Key:      "",
	}

	lifetimeSetting := os.Getenv("JWT_LIFETIME")
	if lifetimeSetting != "" {
		l, err := strconv.ParseInt(lifetimeSetting, 10, 64)
		if err != nil {
			panic(fmt.Errorf("an error occurred while loading jwt lifetime: %v", err))
		}
		jwtConfig.Lifetime = time.Duration(time.Duration(l) * time.Minute)
	}

	issuerSetting := os.Getenv("JWT_ISSUER")
	if issuerSetting != "" {
		jwtConfig.Issuer = issuerSetting
	}

	audienceSetting := os.Getenv("JWT_AUDIENCE")
	if audienceSetting != "" {
		jwtConfig.Audience = strings.Split(audienceSetting, ",")
	}
	slices.Sort(jwtConfig.Audience)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic(fmt.Errorf("JWT_SECRET was not set"))
	} else {
		jwtConfig.Key = jwtSecret
	}

	return jwtConfig
}

func loadPushNotificationConfig() *PushNotificationConfig {
	privateVapidKey := os.Getenv("VAPID_PRIVATE")
	if privateVapidKey == "" {
		panic(fmt.Errorf("VAPID_PRIVATE was not set"))
	}

	publicVapidKey := os.Getenv("VAPID_PUBLIC")
	if publicVapidKey == "" {
		panic(fmt.Errorf("VAPID_PUBLIC was not set"))
	}

	pushNotificationSub := os.Getenv("PUSH_SUB")
	if pushNotificationSub == "" {
		panic(fmt.Errorf("PUSH_SUB was not set"))
	}

	return &PushNotificationConfig{
		VapidPrivateKey: privateVapidKey,
		VapidPublicKey:  publicVapidKey,
		Sub:             pushNotificationSub,
	}
}

func loadWebAuthnConfig() *WebAuthnConfig {
	displayName := os.Getenv("RP_DISPLAY_NAME")
	if displayName == "" {
		panic("RP_DISPLAY_NAME was not set")
	}

	rpId := os.Getenv("RPID")
	if rpId == "" {
		panic("RPID was not set")
	}

	rpOrigins := strings.Split(os.Getenv("RP_ORIGINS"), ",")
	if len(rpOrigins) == 0 {
		panic("RP_ORIGINS was not set")
	}
	for i := range rpOrigins {
		if rpOrigins[i] == "" {
			panic(fmt.Sprintf("Invalid RP_ORIGINS were dectected: %v+", rpOrigins))
		}
	}
	return &WebAuthnConfig{
		RPDisplayName: displayName,
		RPID:          rpId,
		RPOrigins:     rpOrigins,
	}
}
