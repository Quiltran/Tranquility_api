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
	// Auth
	Login(ctx context.Context, cred *models.AuthUser) (*models.AuthUser, error)
	Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)
	RefreshToken(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error)

	// Attachment
	CreateAttachment(ctx context.Context, file *multipart.File, attachment *models.Attachment) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, fileId int32, userId int32) error

	// Guild
	GetJoinedGuilds(ctx context.Context, userId int32) ([]models.Guild, error)
	GetOwnedGuilds(ctx context.Context, userId int32) ([]models.Guild, error)
	GetGuildByID(ctx context.Context, guildId, userId int32) (*models.Guild, error)
}
