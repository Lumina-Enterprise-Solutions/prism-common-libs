package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	BaseModel
	Email        string `json:"email" gorm:"uniqueIndex"` // Seharusnya unik per tenant, bukan global jika multi-tenant
	PasswordHash string `json:"-"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Status       string `json:"status" gorm:"default:active"`
	Roles        []Role `json:"roles" gorm:"many2many:user_roles;"`
}

// PermissionMap mendefinisikan struktur untuk permissions.
// Key: Nama Resource (e.g., "users", "articles")
// Value: Slice dari action yang diizinkan (e.g., "create", "read", "update", "delete")
type PermissionMap map[string][]string

type Role struct {
	BaseModel
	Name        string        `json:"name" gorm:"uniqueIndex"`                       // Seharusnya unik per tenant
	Permissions PermissionMap `json:"permissions" gorm:"type:jsonb;serializer:json"` // [MODIFIKASI]
	Users       []User        `json:"-" gorm:"many2many:user_roles;"`
}

type Tenant struct {
	BaseModel
	Name   string `json:"name"`
	Slug   string `json:"slug" gorm:"uniqueIndex"`
	Status string `json:"status" gorm:"default:active"`
}
