package models

import (
	"time"
	"gorm.io/gorm"
)

// Project represents a project container
type Project struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description" gorm:"default:null"`
	UserID      string         `json:"userId" gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Environments []Environment `json:"environments,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Services     []Service     `json:"services,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}