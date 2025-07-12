package model

import "time"

type Tenant struct {
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Domain    *string   `json:"domain,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
