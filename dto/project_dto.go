package dto

import (
	"github.com/pendeploy-simple/models"
	"time"
)

// ProjectFilter represents filter criteria for projects
type ProjectFilter struct {
	UserID    string
	Search    string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
	IsAdmin   bool
}

// ProjectListResponse represents paginated project list response
type ProjectListResponse struct {
	Projects   []models.Project `json:"projects"`
	TotalCount int64            `json:"totalCount"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// ProjectStatsResponse represents project statistics for dashboard view
type ProjectStatsResponse struct {
	Project struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"createdAt"`
	} `json:"project"`

	Environments struct {
		Total         int                    `json:"total"`
		Environments  []ProjectEnvironmentItem `json:"environments"`
	} `json:"environments"`

	Services struct {
		Total       int                      `json:"total"`
		ByType      map[string]int           `json:"byType"`
		ByStatus    map[string]int           `json:"byStatus"`
		ServiceList []ProjectServiceStatsItem `json:"servicesList"`
	} `json:"services"`

	Deployments struct {
		Total       int64   `json:"total"`
		Successful  int64   `json:"successful"`
		Failed      int64   `json:"failed"`
		InProgress  int64   `json:"inProgress"`
		SuccessRate float64 `json:"successRate"`
	} `json:"deployments"`
}

// ProjectEnvironmentItem represents an environment item in project statistics
type ProjectEnvironmentItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ServicesCount int   `json:"servicesCount"`
	CreatedAt   string `json:"createdAt"`
}

// ProjectServiceStatsItem represents a service item in project statistics
type ProjectServiceStatsItem struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	EnvironmentID string  `json:"environmentId"`
	EnvironmentName string `json:"environmentName"`
	Deployments   int64   `json:"deployments"`
	SuccessRate   float64 `json:"successRate"`
	Replicas      int     `json:"replicas"`
	IsAutoScaling bool    `json:"isAutoScaling"`
}

// CreateProjectRequest represents the request payload for creating a new project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the request payload for updating an existing project
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ProjectResponse represents the standard response format for a project
type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      string    `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
