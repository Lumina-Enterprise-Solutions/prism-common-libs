// file: common/prism-common-libs/model/user.go
package model

import "time"

type User struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id,omitempty"`
	Email        string            `json:"email"`
	PasswordHash string            `json:"-"`
	FirstName    string            `json:"first_name,omitempty"`
	LastName     string            `json:"last_name,omitempty"`
	PhoneNumber  *string           `json:"phone_number,omitempty"`
	AvatarURL    *string           `json:"avatar_url,omitempty"`
	RoleID       string            `json:"-"`
	RoleName     string            `json:"role"`
	Status       string            `json:"status"`
	Is2FAEnabled bool              `json:"is_2fa_enabled"`
	TOTPSecret   string            `json:"-"`
	Bio          *string           `json:"bio,omitempty"`
	JobTitle     *string           `json:"job_title,omitempty"`
	Timezone     *string           `json:"timezone,omitempty"`
	SocialLinks  map[string]string `json:"social_links,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
