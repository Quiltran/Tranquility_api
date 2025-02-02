package config

import (
	"errors"
	"os"
)

type Config struct {
	ConnectionString string
	JWTSecret        string
}

func NewConfig() (*Config, error) {
	connectionString := os.Getenv("CONNECTION_STRING")
	if connectionString == "" {
		return nil, errors.New("CONNECTION_STRING was not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET was not set")
	}

	return &Config{
		connectionString,
		jwtSecret,
	}, nil
}
