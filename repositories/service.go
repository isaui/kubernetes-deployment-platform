package repositories

import (
	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/models"
	"gorm.io/gorm"
)

// ServiceRepository handles database operations for services
type ServiceRepository struct{}

// NewServiceRepository creates a new service repository instance
func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{}
}

// FindAll retrieves all services
func (r *ServiceRepository) FindAll() ([]models.Service, error) {
	var services []models.Service
	result := database.DB.Find(&services)
	return services, result.Error
}

// FindByID retrieves a service by its ID
func (r *ServiceRepository) FindByID(id string) (models.Service, error) {
	var service models.Service
	result := database.DB.First(&service, "id = ?", id)
	return service, result.Error
}

// FindByProjectID retrieves all services belonging to a project
func (r *ServiceRepository) FindByProjectID(projectID string) ([]models.Service, error) {
	var services []models.Service
	result := database.DB.Where("project_id = ?", projectID).Find(&services)
	return services, result.Error
}

// Create inserts a new service into the database
func (r *ServiceRepository) Create(service models.Service) (models.Service, error) {
	result := database.DB.Create(&service)
	return service, result.Error
}

// Update modifies an existing service
func (r *ServiceRepository) Update(service models.Service) error {
	result := database.DB.Save(&service)
	return result.Error
}

// Delete removes a service from the database
func (r *ServiceRepository) Delete(id string) error {
	result := database.DB.Delete(&models.Service{}, "id = ?", id)
	return result.Error
}

// WithDeployments loads service with its deployments
func (r *ServiceRepository) WithDeployments(id string) (models.Service, error) {
	var service models.Service
	result := database.DB.Preload("Deployments").First(&service, "id = ?", id)
	return service, result.Error
}

// CountByProjectID counts services belonging to a project
func (r *ServiceRepository) CountByProjectID(projectID string) (int64, error) {
	var count int64
	result := database.DB.Model(&models.Service{}).Where("project_id = ?", projectID).Count(&count)
	return count, result.Error
}

// UpdateScalingConfig updates service scaling configuration
func (r *ServiceRepository) UpdateScalingConfig(id string, isStatic bool, replicas, minReplicas, maxReplicas int) error {
	return database.DB.Model(&models.Service{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_static_replica": isStatic,
			"replicas":          replicas,
			"min_replicas":      minReplicas,
			"max_replicas":      maxReplicas,
		}).Error
}

// DB returns the database instance
func (r *ServiceRepository) DB() *gorm.DB {
	return database.DB
}
