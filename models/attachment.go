package models

import "time"

type Attachment struct {
	ID           int       `json:"id,omitempty" db:"id"`
	FileName     string    `json:"file_name,omitempty" db:"file_name"`
	FilePath     string    `json:"file_path,omitempty" db:"file_path"`
	FileSize     int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType     string    `json:"mime_type,omitempty" db:"mime_type"`
	UserUploaded int32     `json:"user_uploaded,omitempty" db:"user_uploaded"`
	CreatedDate  time.Time `json:"created_date,omitempty" db:"created_date"`
}
