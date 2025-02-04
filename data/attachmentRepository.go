package data

import (
	"context"
	"tranquility/models"

	"github.com/jmoiron/sqlx"
)

type attachmentRepo struct {
	db *sqlx.DB
}

func (a *attachmentRepo) CreateAttachment(ctx context.Context, attachment *models.Attachment) (*models.Attachment, error) {
	var output models.Attachment
	err := a.db.QueryRowxContext(
		ctx,
		`INSERT INTO attachment (file_name, file_path, file_size, mime_type, user_uploaded)
		SELECT $1, $2, $3, $4, $5
		RETURNING id, file_name, file_path, file_size, mime_type, created_date;`,
		&attachment.FileName,
		&attachment.FilePath,
		&attachment.FileSize,
		&attachment.MimeType,
		&attachment.UserUploaded,
	).StructScan(&output)
	return &output, err
}
