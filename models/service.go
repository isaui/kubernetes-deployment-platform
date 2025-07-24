// models/service.go  
package models

import (
	"time"
	"gorm.io/gorm"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// EnvVars custom type for JSON storage
type EnvVars map[string]string

func (e EnvVars) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EnvVars) Scan(value interface{}) error {
	if value == nil {
		*e = make(map[string]string)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, e)
}

// ServiceType represents different service types
type ServiceType string

const (
	ServiceTypeGit     ServiceType = "git"      // Git-based applications (web, workers, etc.)
	ServiceTypeManaged ServiceType = "managed"  // Managed services (databases, cache, storage, etc.)
)

// Service represents a deployable service
type Service struct {
	// Common fields for all service types
	ID            string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string         `json:"name" gorm:"not null"`
	Type          ServiceType    `json:"type" gorm:"type:varchar(20);default:'git'"`
	ProjectID     string         `json:"projectId" gorm:"type:uuid;not null;index"`
	
	// Git repository (only applicable for ServiceTypeGit)
	RepoURL       string         `json:"repoUrl" gorm:"default:null"`
	Branch        string         `json:"branch" gorm:"default:main"`
	
	// Managed services specific fields (only applicable for ServiceTypeManaged)
	ManagedType   string         `json:"managedType" gorm:"default:null"` // postgresql, redis, minio, etc.
	Version       string         `json:"version" gorm:"default:null"`     // 14, 6.0, latest, etc.
	StorageSize   string         `json:"storageSize" gorm:"default:null"` // 1Gi, 10Gi, etc.
	
	// Environment reference
	EnvironmentID string         `json:"environmentId" gorm:"type:uuid;index"`
	
	// Deployment config (all in one place)
	Port          int            `json:"port" gorm:"default:3000"`
	EnvVars       EnvVars        `json:"envVars" gorm:"type:jsonb;default:'{}'"`
	BuildCommand  string         `json:"buildCommand" gorm:"default:null"`
	StartCommand  string         `json:"startCommand" gorm:"default:null"`
	
	// Resources & Scaling
	CPULimit      string         `json:"cpuLimit" gorm:"default:1024m"`
	MemoryLimit   string         `json:"memoryLimit" gorm:"default:2Gi"`
	IsStaticReplica bool          `json:"isStaticReplica" gorm:"default:true"`
	Replicas      int            `json:"replicas" gorm:"default:1"`
	MinReplicas   int            `json:"minReplicas" gorm:"default:1"`
	MaxReplicas   int            `json:"maxReplicas" gorm:"default:3"`
	
	// Domain
	Domain        string         `json:"domain" gorm:"default:null"` // auto-generated
	CustomDomain  string         `json:"customDomain" gorm:"default:null"`
	
	// Status
	Status        string         `json:"status" gorm:"default:inactive"` // inactive, building, running, failed
	
	// API Key for webhooks
	APIKey        string         `json:"apiKey" gorm:"type:uuid;default:gen_random_uuid()"`
	
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Project     Project      `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Environment Environment   `json:"environment,omitempty" gorm:"foreignKey:EnvironmentID"`
	Deployments []Deployment `json:"deployments,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
}