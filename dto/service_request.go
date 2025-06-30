package dto

import "github.com/pendeploy-simple/models"

// ServiceRequest represents a service creation/update request - UPDATED untuk managed services
type ServiceRequest struct {
	// Common fields for all service types
	Name          string             `json:"name" binding:"required"`
	Type          models.ServiceType `json:"type" binding:"required"` // "git" or "managed"
	ProjectID     string             `json:"projectId" binding:"required"`
	EnvironmentID string             `json:"environmentId" binding:"required"`
	
	// Git-specific fields (required only when Type is "git")
	RepoURL       string             `json:"repoUrl"`
	Branch        string             `json:"branch"`
	Port          int                `json:"port"`
	BuildCommand  string             `json:"buildCommand"`
	StartCommand  string             `json:"startCommand"`
	
	// Managed service specific fields (required only when Type is "managed")
	ManagedType   string             `json:"managedType"` // postgresql, redis, minio, etc.
	Version       string             `json:"version"`     // 14, 6.0, latest, etc.
	StorageSize   string             `json:"storageSize"` // 1Gi, 10Gi, etc.
	
	// Common configuration fields
	EnvVars       models.EnvVars     `json:"envVars"`
	CPULimit      string             `json:"cpuLimit"`
	MemoryLimit   string             `json:"memoryLimit"`
	IsStaticReplica bool             `json:"isStaticReplica"`
	Replicas      int                `json:"replicas"`
	MinReplicas   int                `json:"minReplicas"`
	MaxReplicas   int                `json:"maxReplicas"`
	CustomDomain  string             `json:"customDomain"`
}