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
	
	// Git info
	CommitSHA     string            `json:"commitSha" gorm:"not null"`
	CommitMessage string            `json:"commitMessage" gorm:"default:null"`
	
	// Build info
	Status        DeploymentStatus  `json:"status" gorm:"type:varchar(20);default:'building'"`
	ImageTag      string            `json:"imageTag" gorm:"not null"` // untuk K8s deployment
	BuildLogs     string            `json:"buildLogs" gorm:"type:text;default:null"`
	
	// Timestamps
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	DeployedAt    *time.Time        `json:"deployedAt" gorm:"default:null"`
	
	// Relation
	Service       Service           `json:"service,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
}