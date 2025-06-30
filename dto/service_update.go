// dto/service_update.go - UPDATED untuk managed services
package dto

import (
	"fmt"
	"github.com/pendeploy-simple/models"
)

// BaseServiceUpdateRequest berisi field umum yang boleh diupdate untuk semua jenis service
type BaseServiceUpdateRequest struct {
	Name          string           `json:"name,omitempty"` 
	EnvVars       models.EnvVars   `json:"envVars,omitempty"`    // Hanya untuk git services
	CPULimit      string           `json:"cpuLimit,omitempty"`
	MemoryLimit   string           `json:"memoryLimit,omitempty"`
	IsStaticReplica *bool          `json:"isStaticReplica,omitempty"`
	Replicas      *int             `json:"replicas,omitempty"`
	MinReplicas   *int             `json:"minReplicas,omitempty"`
	MaxReplicas   *int             `json:"maxReplicas,omitempty"`
	CustomDomain  string           `json:"customDomain,omitempty"`
}

// GitServiceUpdateRequest berisi field yang boleh diupdate untuk service bertipe git
type GitServiceUpdateRequest struct {
	BaseServiceUpdateRequest
	Branch        string           `json:"branch,omitempty"`
	Port          *int             `json:"port,omitempty"`
	BuildCommand  string           `json:"buildCommand,omitempty"`
	StartCommand  string           `json:"startCommand,omitempty"`
}

// ManagedServiceUpdateRequest berisi field yang boleh diupdate untuk service bertipe managed
type ManagedServiceUpdateRequest struct {
	BaseServiceUpdateRequest
	Version       string           `json:"version,omitempty"`
	StorageSize   string           `json:"storageSize,omitempty"`
}

// ServiceUpdateRequest adalah wrapper untuk request update service
// Type digunakan untuk menentukan apakah ini update untuk git service atau managed service
type ServiceUpdateRequest struct {
	Type string `json:"type" binding:"required,oneof=git managed"`
	Git  *GitServiceUpdateRequest `json:"git,omitempty"`
	Managed *ManagedServiceUpdateRequest `json:"managed,omitempty"`
}

// UpdateServiceModel mengupdate model service berdasarkan request yang diterima - UPDATED untuk managed services
// Pengecekan tipe service dan validasi dilakukan diluar fungsi ini
func (req *ServiceUpdateRequest) UpdateServiceModel(service *models.Service) {
	var base BaseServiceUpdateRequest
	
	// Tentukan base request berdasarkan tipe
	if req.Type == "git" && req.Git != nil {
		base = req.Git.BaseServiceUpdateRequest
	} else if req.Type == "managed" && req.Managed != nil {
		base = req.Managed.BaseServiceUpdateRequest
	}
	
	// Update common fields jika disediakan
	if base.Name != "" {
		service.Name = base.Name
	}
	
	// EnvironmentID tidak boleh diubah jadi tidak diproses disini
	
	// EnvVars hanya boleh diupdate untuk git services, managed services auto-generated
	if req.Type == "git" && base.EnvVars != nil {
		service.EnvVars = base.EnvVars
	}
	// Untuk managed services, EnvVars diabaikan karena auto-generated
	
	if base.CPULimit != "" {
		service.CPULimit = base.CPULimit
	}
	
	if base.MemoryLimit != "" {
		service.MemoryLimit = base.MemoryLimit
	}
	
	// Scaling configuration - managed services biasanya single replica tapi bisa di-override
	if base.IsStaticReplica != nil {
		if req.Type == "git" {
			service.IsStaticReplica = *base.IsStaticReplica
		}
		// Untuk managed services, biasanya static replica = true, tapi bisa diubah untuk beberapa services
	}
	
	if base.Replicas != nil {
		if req.Type == "git" {
			service.Replicas = *base.Replicas
		} else if req.Type == "managed" {
			// Managed services biasanya single replica, tapi allow update untuk services tertentu
			service.Replicas = *base.Replicas
		}
	}
	
	if base.MinReplicas != nil {
		if req.Type == "git" {
			service.MinReplicas = *base.MinReplicas
		}
		// Managed services jarang pakai autoscaling, tapi bisa untuk services tertentu
	}
	
	if base.MaxReplicas != nil {
		if req.Type == "git" {
			service.MaxReplicas = *base.MaxReplicas
		}
		// Managed services jarang pakai autoscaling, tapi bisa untuk services tertentu
	}
	
	if base.CustomDomain != "" {
		service.CustomDomain = base.CustomDomain
	}
	
	// Update type-specific fields jika disediakan
	if req.Type == "git" && req.Git != nil {
		if req.Git.Branch != "" {
			service.Branch = req.Git.Branch
		}
		
		if req.Git.Port != nil {
			service.Port = *req.Git.Port
		}
		
		if req.Git.BuildCommand != "" {
			service.BuildCommand = req.Git.BuildCommand
		}
		
		if req.Git.StartCommand != "" {
			service.StartCommand = req.Git.StartCommand
		}
	} else if req.Type == "managed" && req.Managed != nil {
		if req.Managed.Version != "" {
			service.Version = req.Managed.Version
		}
		
		if req.Managed.StorageSize != "" {
			service.StorageSize = req.Managed.StorageSize
		}
	}
}

// ValidateServiceUpdateRequest validates service update request based on type
func (req *ServiceUpdateRequest) ValidateServiceUpdateRequest() error {
	if req.Type == "git" {
		if req.Git == nil {
			return fmt.Errorf("git configuration is required for git service updates")
		}
		if req.Managed != nil {
			return fmt.Errorf("managed configuration should not be provided for git services")
		}
	} else if req.Type == "managed" {
		if req.Managed == nil {
			return fmt.Errorf("managed configuration is required for managed service updates")
		}
		if req.Git != nil {
			return fmt.Errorf("git configuration should not be provided for managed services")
		}
		
		// Validate managed service specific fields
		if req.Managed.EnvVars != nil && len(req.Managed.EnvVars) > 0 {
			return fmt.Errorf("environment variables are auto-generated for managed services and cannot be modified")
		}
		
		// Validate storage size format if provided
		if req.Managed.StorageSize != "" && len(req.Managed.StorageSize) < 2 {
			return fmt.Errorf("invalid storage size format")
		}
	}
	
	return nil
}