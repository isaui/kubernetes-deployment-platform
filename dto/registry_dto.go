package dto

import (
	"time"
	
	"github.com/pendeploy-simple/models"
)

// RegistryFilter represents filter criteria for registries
type RegistryFilter struct {
	Search     string
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
	OnlyActive bool
}

// RegistryResponse represents the response format for a registry
type RegistryResponse struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	URL       string             `json:"url"`
	IsDefault bool               `json:"isDefault"`
	IsActive  bool               `json:"isActive"`
	Status    models.RegistryStatus `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// RegistryListResponse represents paginated registry list response
type RegistryListResponse struct {
	Registries []RegistryResponse `json:"registries"`
	TotalCount int64              `json:"totalCount"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

// CreateRegistryRequest represents the request payload for creating a new registry
type CreateRegistryRequest struct {
	Name      string `json:"name" binding:"required"`
	Password  string `json:"password"` // Optional - will be auto-generated if not provided
	IsDefault bool   `json:"isDefault"`
}

// UpdateRegistryRequest represents the request payload for updating an existing registry
type UpdateRegistryRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	IsDefault bool   `json:"isDefault"`
	IsActive  bool   `json:"isActive"`
}

// RegistryCredentials holds the access information for a registry
type RegistryCredentials struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // Only returned on initial creation
}

// RegistryDetailsResponse represents detailed information for a single registry including Kubernetes info
type RegistryDetailsResponse struct {
	Registry    RegistryResponse `json:"registry"`
	Credentials *RegistryCredentials `json:"credentials,omitempty"`
	Images      int              `json:"images"`
	Size        int64            `json:"size"` // in bytes
	IsHealthy   bool             `json:"isHealthy"`
	KubeStatus  string           `json:"kubeStatus"`
	LastSynced  *time.Time       `json:"lastSynced"`
}

// RegistryImageInfo represents information about an image in the registry
type RegistryImageInfo struct {
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// RegistryImagesResponse represents the response containing registry images
type RegistryImagesResponse struct {
	Images     []RegistryImageInfo `json:"images"`
	TotalCount int                 `json:"totalCount"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
}
