package models

import "time"

type AttachmentResponse struct {
	ID          int       `json:"id,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	FilePath    string    `json:"file_path,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	MimeType    string    `json:"mime_type,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
}
