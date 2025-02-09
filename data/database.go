package data

import (
	"context"
	"errors"
	"mime/multipart"
	"tranquility/models"
)

var (
	ErrMissingPassword    = errors.New("password is required")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// This interface is used when creating new controllers.
type IDatabase interface {
	Login(ctx context.Context, cred *models.AuthUser) (*models.AuthUser, error)
	Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)
	RefreshToken(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)
	CreateAttachment(ctx context.Context, file *multipart.File, attachment *models.Attachment) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, fileId int32, userId int32) error
}
