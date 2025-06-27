// models/deployment.go
package models

import (
	"time"
)

// DeploymentStatus represents deployment status
type DeploymentStatus string

const (
	DeploymentStatusBuilding  DeploymentStatus = "building"
	DeploymentStatusSuccess   DeploymentStatus = "success"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)

// Deployment represents a deployment instance
type Deployment struct {
	ID            string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ServiceID     string            `json:"serviceId" gorm:"type:uuid;not null;index"`
	
	// Git info - optional for managed services
	CommitSHA     string            `json:"commitSha" gorm:"default:null"`
	CommitMessage string            `json:"commitMessage" gorm:"default:null"`
	
	// Build info
	Status        DeploymentStatus  `json:"status" gorm:"type:varchar(20);default:'building'"`
	Image         string            `json:"image" gorm:"default:null"` // optional for managed services
	// Managed service specific
	Version       string            `json:"version" gorm:"type:varchar(50);default:null"` // For tracking version changes in managed services
	
	// Timestamps
	CreatedAt     time.Time         `json:"createdAt" gorm:"autoCreateTime"`
	
	// Relation
	Service       Service           `json:"service,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
}