package data

import (
	"context"
	"database/sql"
	"fmt"
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

func (a *attachmentRepo) DeleteAttachment(ctx context.Context, fileId, userId int32) (*sql.Tx, string, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", err
	}

	var fileName string
	rows, err := tx.QueryContext(
		ctx,
		`DELETE FROM attachment WHERE id = $1 and user_uploaded = $2 RETURNING file_name;`,
		fileId,
		userId,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	if !rows.Next() {
		tx.Rollback()
		return nil, "", nil
	}

	if err := rows.Scan(&fileName); err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("an error occurred while scanning file name while deleting: %v", err)
	}

	if rows.Next() {
		tx.Rollback()
		return nil, "", fmt.Errorf("more than one file was affected by the delete operation on attachment %d", fileId)
	}

	return tx, fileName, nil
}
