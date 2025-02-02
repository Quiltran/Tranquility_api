package services

import (
	"context"
	"tranquility/data"
	"tranquility/models"
)

type DatabaseCommands struct {
	db data.IDatabase
}

func NewDatabaseCommands(db data.IDatabase) *DatabaseCommands {
	return &DatabaseCommands{db}
}

func (d *DatabaseCommands) Login(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	return d.db.Login(ctx, user)
}

func (d *DatabaseCommands) Register(ctx context.Context, user *models.AuthUser) (*models.AuthUser, error) {
	return d.db.Register(ctx, user)
}
