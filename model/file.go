package model

import "time"

type FileMetadata struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	StoragePath  string    `json:"-"` // Sembunyikan dari klien
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	OwnerUserID  *string   `json:"owner_user_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
