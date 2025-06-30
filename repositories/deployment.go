package repositories

import (
	"time"

	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/models"
	"gorm.io/gorm"
)

// DeploymentRepository handles database operations for deployments
type DeploymentRepository struct{}

// NewDeploymentRepository creates a new deployment repository instance
func NewDeploymentRepository() *DeploymentRepository {
	return &DeploymentRepository{}
}

// FindAll retrieves all deployments
func (r *DeploymentRepository) FindAll() ([]models.Deployment, error) {
	var deployments []models.Deployment
	result := database.DB.Find(&deployments)
	return deployments, result.Error
}

// FindByID retrieves a deployment by its ID
func (r *DeploymentRepository) FindByID(id string) (models.Deployment, error) {
	var deployment models.Deployment
	result := database.DB.First(&deployment, "id = ?", id)
	return deployment, result.Error
}

// FindByServiceID retrieves all deployments for a service
func (r *DeploymentRepository) FindByServiceID(serviceID string) ([]models.Deployment, error) {
	var deployments []models.Deployment
	result := database.DB.Where("service_id = ?", serviceID).Order("created_at DESC").Find(&deployments)
	return deployments, result.Error
}

func (r *DeploymentRepository) UpdateImage(id string, image string) error {
	var updates = map[string]interface{}{
		"image": image,
	}
	result := database.DB.Model(&models.Deployment{}).
		Where("id = ?", id).
		Updates(updates)
	return result.Error
}

// Create inserts a new deployment into the database
func (r *DeploymentRepository) Create(deployment models.Deployment) (models.Deployment, error) {
	result := database.DB.Create(&deployment)
	return deployment, result.Error
}

// UpdateStatus updates the status of a deployment
func (r *DeploymentRepository) UpdateStatus(id string, status models.DeploymentStatus) error {
	var updates = map[string]interface{}{
		"status": status,
	}
	
	// If status is success, set the deployed_at timestamp
	if status == models.DeploymentStatusSuccess {
		now := time.Now()
		updates["deployed_at"] = &now
	}
	
	result := database.DB.Model(&models.Deployment{}).
		Where("id = ?", id).
		Updates(updates)
		
	return result.Error
}


// GetLatestSuccessfulDeployment retrieves the most recent successful deployment for a service
func (r *DeploymentRepository) GetLatestSuccessfulDeployment(serviceID string) (models.Deployment, error) {
	var deployment models.Deployment
	result := database.DB.Where("service_id = ? AND status = ?", 
		serviceID, models.DeploymentStatusSuccess).
		Order("created_at DESC").First(&deployment)
	return deployment, result.Error
}

func (r *DeploymentRepository) GetLatestDeployment(serviceID string) (models.Deployment, error) {
	var deployment models.Deployment
	result := database.DB.Where("service_id = ?", serviceID).
		Order("created_at DESC").First(&deployment)
	return deployment, result.Error
}

// CountByServiceID counts the number of deployments for a service
func (r *DeploymentRepository) CountByServiceID(serviceID string) (int64, error) {
	var count int64
	result := database.DB.Model(&models.Deployment{}).
		Where("service_id = ?", serviceID).Count(&count)
	return count, result.Error
}

// GetSuccessRate calculates the deployment success rate for a service
func (r *DeploymentRepository) GetSuccessRate(serviceID string) (float64, error) {
	type Result struct {
		Total       int64
		Successful  int64
	}
	
	var result Result
	
	err := database.DB.Raw(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as successful
		FROM 
			deployments
		WHERE 
			service_id = ?
			AND deleted_at IS NULL
	`, models.DeploymentStatusSuccess, serviceID).Scan(&result).Error
	
	if err != nil {
		return 0, err
	}
	
	if result.Total == 0 {
		return 0, nil
	}
	
	return float64(result.Successful) / float64(result.Total), nil
}

// FindByProjectID retrieves all deployments for services in a project
func (r *DeploymentRepository) FindByProjectID(projectID string) ([]models.Deployment, error) {
	var deployments []models.Deployment
	
	// Join with services to filter by project ID
	result := database.DB.Joins("JOIN services ON services.id = deployments.service_id").
		Where("services.project_id = ?", projectID).
		Order("deployments.created_at DESC").
		Find(&deployments)
		
	return deployments, result.Error
}

// CountByProjectID counts the total number of deployments for all services in a project
func (r *DeploymentRepository) CountByProjectID(projectID string) (int64, error) {
	var count int64
	
	// Join with services to filter by project ID
	result := database.DB.Model(&models.Deployment{}).
		Joins("JOIN services ON services.id = deployments.service_id").
		Where("services.project_id = ?", projectID).
		Count(&count)
		
	return count, result.Error
}

// CountDeploymentsByProjectIDAndStatus counts the number of deployments for a project with a specific status
func (r *DeploymentRepository) CountDeploymentsByProjectIDAndStatus(projectID string, status models.DeploymentStatus) (int64, error) {
	var count int64
	
	// Join with services to filter by project ID
	result := database.DB.Model(&models.Deployment{}).
		Joins("JOIN services ON services.id = deployments.service_id").
		Where("services.project_id = ? AND deployments.status = ?", projectID, status).
		Count(&count)
		
	return count, result.Error
}

// DB returns the database instance
func (r *DeploymentRepository) DB() *gorm.DB {
	return database.DB
}
