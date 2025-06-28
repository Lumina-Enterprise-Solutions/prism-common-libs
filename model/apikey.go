package model

import "time"

// APIKeyMetadata menyimpan informasi tentang sebuah API key, tanpa menyertakan key rahasianya.
// Ini aman untuk ditampilkan ke pengguna.
type APIKeyMetadata struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Prefix      string     `json:"prefix"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}
