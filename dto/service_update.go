package dto

import (
	"github.com/pendeploy-simple/models"
)

// BaseServiceUpdateRequest berisi field umum yang boleh diupdate untuk semua jenis service
type BaseServiceUpdateRequest struct {
	Name          string           `json:"name,omitempty"` 
	EnvVars       models.EnvVars   `json:"envVars,omitempty"`
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

// UpdateServiceModel mengupdate model service berdasarkan request yang diterima
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
	
	if base.EnvVars != nil {
		service.EnvVars = base.EnvVars
	}
	
	if base.CPULimit != "" {
		service.CPULimit = base.CPULimit
	}
	
	if base.MemoryLimit != "" {
		service.MemoryLimit = base.MemoryLimit
	}
	
	if base.IsStaticReplica != nil {
		service.IsStaticReplica = *base.IsStaticReplica
	}
	
	if base.Replicas != nil {
		service.Replicas = *base.Replicas
	}
	
	if base.MinReplicas != nil {
		service.MinReplicas = *base.MinReplicas
	}
	
	if base.MaxReplicas != nil {
		service.MaxReplicas = *base.MaxReplicas
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
