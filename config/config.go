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
	*JWTConfig
}

type JWTConfig struct {
	Lifetime time.Duration
	Issuer   string
	Audience []string
	Key      string
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

	return &Config{
		connectionString,
		uploadPath,
		jwtConfig,
	}, nil
}
