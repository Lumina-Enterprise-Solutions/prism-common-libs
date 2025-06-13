// file: common/prism-common-libs/model/user.go
package model

import "time"

// User adalah Single Source of Truth untuk model pengguna.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	PhoneNumber  *string   `json:"phone_number,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Is2FAEnabled bool      `json:"is_2fa_enabled"` // <-- TAMBAHKAN
	TOTPSecret   string    `json:"-"`              // <-- TAMBAHKAN (sembunyikan dari JSON)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
