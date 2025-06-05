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
	Email        string `json:"email" gorm:"index"`
	PasswordHash string `json:"-"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Status       string `json:"status" gorm:"default:active"`
	TenantID     string `json:"tenant_id" gorm:"index;not null"`                                                                           // User sekarang jelas terikat ke satu tenant
	Roles        []Role `json:"roles" gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"` // Eksplisit foreign keys

	// Kolom untuk integrasi AD
	ADUserPrincipalName string     `json:"ad_user_principal_name,omitempty" gorm:"index"`
	ADObjectID          string     `json:"ad_object_id,omitempty" gorm:"index;unique"` // Seharusnya unik global
	IsADManaged         bool       `json:"is_ad_managed,omitempty" gorm:"default:false"`
	LastADSync          *time.Time `json:"last_ad_sync,omitempty"`
}

// PermissionMap mendefinisikan struktur untuk permissions.
// Key: Nama Resource (e.g., "users", "articles")
// Value: Slice dari action yang diizinkan (e.g., "create", "read", "update", "delete")
type PermissionMap map[string][]string

type Role struct {
	BaseModel
	Name        string        `json:"name" gorm:"uniqueIndex:idx_role_name_tenant_id"` // Unique per tenant
	TenantID    string        `json:"tenant_id" gorm:"uniqueIndex:idx_role_name_tenant_id;not null"`
	Permissions PermissionMap `json:"permissions" gorm:"type:jsonb;serializer:json"`
	Users       []User        `json:"-" gorm:"many2many:user_roles;"` // user_roles akan ada di skema tenant
}

type Tenant struct {
	BaseModel
	Name   string `json:"name"`
	Slug   string `json:"slug" gorm:"uniqueIndex"`
	Status string `json:"status" gorm:"default:active"`
}

// Model untuk ADGroupRoleMapping
type ADGroupRoleMapping struct {
	BaseModel
	ADGroupName string    `json:"ad_group_name" gorm:"not null;uniqueIndex:idx_mapping_tenant_adgroup"`
	RoleID      uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	// Role        Role      `json:"role,omitempty" gorm:"foreignKey:RoleID"` // Relasi ini mungkin sulit jika Role per tenant dan mapping di public
	TenantID string `json:"tenant_id" gorm:"not null;uniqueIndex:idx_mapping_tenant_adgroup"`
}
