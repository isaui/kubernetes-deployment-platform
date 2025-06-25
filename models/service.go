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
	ServiceTypeWeb      ServiceType = "web"
	ServiceTypeWorker   ServiceType = "worker"  
	ServiceTypeDatabase ServiceType = "database"
)

// Service represents a deployable service
type Service struct {
	ID            string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string         `json:"name" gorm:"not null"`
	Type          ServiceType    `json:"type" gorm:"type:varchar(20);default:'web'"`
	ProjectID     string         `json:"projectId" gorm:"type:uuid;not null;index"`
	
	// Git repository
	RepoURL       string         `json:"repoUrl" gorm:"not null"`
	Branch        string         `json:"branch" gorm:"default:main"`
	
	// Environment reference
	EnvironmentID string         `json:"environmentId" gorm:"type:uuid;index"`
	
	// Deployment config (all in one place)
	Port          int            `json:"port" gorm:"default:3000"`
	EnvVars       EnvVars        `json:"envVars" gorm:"type:jsonb;default:'{}'"`
	BuildCommand  string         `json:"buildCommand" gorm:"default:null"`
	StartCommand  string         `json:"startCommand" gorm:"default:null"`
	
	// Resources & Scaling
	CPULimit      string         `json:"cpuLimit" gorm:"default:500m"`
	MemoryLimit   string         `json:"memoryLimit" gorm:"default:512Mi"`
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