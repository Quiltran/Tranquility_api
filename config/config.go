package config

import (
	"errors"
	"os"
)

type Config struct {
	ConnectionString string
	UploadPath       string
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

	return &Config{
		connectionString,
		uploadPath,
	}, nil
}
