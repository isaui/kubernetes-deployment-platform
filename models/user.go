package models

import (
	"time"
	"gorm.io/gorm"
)

// Role represents user role types
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a user in the system
type User struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Password is not exposed in JSON
	Username  *string        `json:"username" gorm:"default:null;uniqueIndex"`
	Name      *string        `json:"name" gorm:"default:null"`
	Role      Role           `json:"role" gorm:"type:varchar(10);default:'user'"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
