package config

import (
	"errors"
	"os"
)

type Config struct {
	ConnectionString string
}

func NewConfig() (*Config, error) {
	connectionString := os.Getenv("CONNECTION_STRING")
	if connectionString == "" {
		return nil, errors.New("CONNECTION_STRING was not set")
	}

	return &Config{
		connectionString,
	}, nil
}
