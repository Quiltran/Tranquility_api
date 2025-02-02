package data

import (
	"context"
	"tranquility/models"
)

// This interface is used when creating new controllers.
type IDatabase interface {
	Login(ctx context.Context, cred *models.AuthUser) (*models.AuthUser, error)
	Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)
}
