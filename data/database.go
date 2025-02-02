package data

import (
	"context"
	"tranquility/models"
)

type IDatabase interface {
	Login(ctx context.Context, cred *models.AuthUser) (*models.AuthUser, error)
	Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)
}
