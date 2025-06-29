package models

import (
	"time"
)

// RegistryStatus represents the status of a registry deployment
type RegistryStatus string

const (
	RegistryStatusPending  RegistryStatus = "pending"
	RegistryStatusBuilding RegistryStatus = "building"
	RegistryStatusReady    RegistryStatus = "ready"
	RegistryStatusFailed   RegistryStatus = "failed"
)

// Registry represents a container registry configuration
type Registry struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"not null"`
	URL          string         `json:"url" gorm:"default:null"`
	IsDefault    bool           `json:"isDefault" gorm:"default:false"`
	IsActive     bool           `json:"isActive" gorm:"default:true"`
	Status       RegistryStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	BuildPodName string         `json:"-" gorm:"default:null"` // Name of the K8s pod handling the build
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}
