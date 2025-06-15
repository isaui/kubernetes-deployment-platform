package models

import (
	"time"
)

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	StatusPending   DeploymentStatus = "pending"
	StatusRunning   DeploymentStatus = "running"
	StatusSucceeded DeploymentStatus = "succeeded"
	StatusFailed    DeploymentStatus = "failed"
)

// Deployment represents a deployment in the system
type Deployment struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	RepoURL     string          `json:"repo_url" binding:"required"`
	Branch      string          `json:"branch" binding:"required"`
	Environment string          `json:"environment" binding:"required"`
	Status      DeploymentStatus `json:"status"`
	Logs        string          `json:"logs" gorm:"type:text"`
	UserID      uint            `json:"user_id" binding:"required"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// Struct untuk request deployment dipindahkan ke deployment_request.go
