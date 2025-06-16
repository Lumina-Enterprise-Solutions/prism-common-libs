package model

import "time"

type UserPreference struct {
	UserID    string    `json:"-"` // Tidak perlu di JSON response
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
