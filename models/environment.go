package models

import (
	"time"
	"gorm.io/gorm"
)

// Environment represents a deployment environment for a project
type Environment struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"` // Name must be unique per project
	Description string         `json:"description" gorm:"default:null"` // Optional description
	ProjectID   string         `json:"projectId" gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Services  []Service `json:"services,omitempty" gorm:"foreignKey:EnvironmentID;constraint:OnDelete:SET NULL"`
}

// TableName sets the table name for Environment model
func (Environment) TableName() string {
	return "environments"
}
