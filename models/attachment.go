package models

import (
	"errors"
	"mime/multipart"
	"net/http"
	"time"
)

var (
	ErrAttachmentNoContentType = errors.New("no content type was provided for the attachment")
	ErrAttachmentNoFileName    = errors.New("no file name was provided for the file")
)

type Attachment struct {
	ID           int       `json:"id,omitempty" db:"id"`
	FileName     string    `json:"file_name,omitempty" db:"file_name"`
	FilePath     string    `json:"file_path,omitempty" db:"file_path"`
	FileSize     int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType     string    `json:"mime_type,omitempty" db:"mime_type"`
	UserUploaded int32     `json:"user_uploaded,omitempty" db:"user_uploaded"`
	CreatedDate  time.Time `json:"created_date,omitempty" db:"created_date"`
}

func NewAttachmentFromRequest(r *http.Request, userId int32, fieldName string) (*Attachment, multipart.File, error) {
	file, handler, err := r.FormFile(fieldName)
	if err != nil {
		return nil, nil, err
	}

	fileType := handler.Header.Get("Content-Type")
	if fileType == "" {
		return nil, nil, ErrAttachmentNoContentType
	}

	fileName := handler.Filename
	if fileName == "" {
		return nil, nil, ErrAttachmentNoFileName
	}

	return &Attachment{
		FileName:     fileName,
		FileSize:     handler.Size,
		MimeType:     fileType,
		UserUploaded: userId,
	}, file, nil
}
