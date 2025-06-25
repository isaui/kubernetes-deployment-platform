package dto

import (
	"time"
)

// EnvironmentRequest is the structure for environment creation/update requests
type EnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ProjectID   string `json:"projectId" binding:"required"`
}

// EnvironmentResponse is the structure for environment responses
type EnvironmentResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectID   string    `json:"projectId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EnvironmentListResponse wraps a list of environments
type EnvironmentListResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
}
