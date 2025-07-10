package model

import "time"

type Team struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id,omitempty"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Department     *string   `json:"department,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
