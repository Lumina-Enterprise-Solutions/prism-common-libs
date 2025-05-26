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
	Email        string `json:"email" gorm:"uniqueIndex"`
	PasswordHash string `json:"-"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Status       string `json:"status" gorm:"default:active"`
	Roles        []Role `json:"roles" gorm:"many2many:user_roles;"`
}

type Role struct {
	BaseModel
	Name        string                 `json:"name"`
	Permissions map[string]interface{} `json:"permissions" gorm:"type:jsonb;serializer:json"`
	Users       []User                 `json:"-" gorm:"many2many:user_roles;"`
}

type Tenant struct {
	BaseModel
	Name   string `json:"name"`
	Slug   string `json:"slug" gorm:"uniqueIndex"`
	Status string `json:"status" gorm:"default:active"`
}
