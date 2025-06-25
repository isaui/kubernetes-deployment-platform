package repositories

import (
	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/models"
)

// EnvironmentRepository handles database operations for environments
type EnvironmentRepository struct{}

// NewEnvironmentRepository creates a new environment repository instance
func NewEnvironmentRepository() *EnvironmentRepository {
	return &EnvironmentRepository{}
}

// FindByID retrieves an environment by its ID
func (r *EnvironmentRepository) FindByID(id string) (models.Environment, error) {
	var environment models.Environment
	result := database.DB.First(&environment, "id = ?", id)
	return environment, result.Error
}

// FindByProjectID retrieves all environments for a project
func (r *EnvironmentRepository) FindByProjectID(projectID string) ([]models.Environment, error) {
	var environments []models.Environment
	result := database.DB.Where("project_id = ?", projectID).Find(&environments)
	return environments, result.Error
}

// CountByProjectID counts the number of environments for a project
func (r *EnvironmentRepository) CountByProjectID(projectID string) (int64, error) {
	var count int64
	result := database.DB.Model(&models.Environment{}).Where("project_id = ?", projectID).Count(&count)
	return count, result.Error
}

// Create inserts a new environment into the database
func (r *EnvironmentRepository) Create(environment models.Environment) (models.Environment, error) {
	result := database.DB.Create(&environment)
	return environment, result.Error
}

// Update modifies an existing environment
func (r *EnvironmentRepository) Update(environment models.Environment) error {
	result := database.DB.Save(&environment)
	return result.Error
}

// ExistsByNameAndProject checks if an environment with the given name exists in a project
func (r *EnvironmentRepository) ExistsByNameAndProject(name string, projectID string) (bool, error) {
	var count int64
	result := database.DB.Model(&models.Environment{}).Where("name = ? AND project_id = ?", name, projectID).Count(&count)
	return count > 0, result.Error
}

// Delete removes an environment from the database (soft delete)
func (r *EnvironmentRepository) Delete(id string) error {
	result := database.DB.Delete(&models.Environment{}, "id = ?", id)
	return result.Error
}

// CountServicesInEnvironment counts the number of services in an environment
func (r *EnvironmentRepository) CountServicesInEnvironment(environmentID string) (int, error) {
	var count int64
	result := database.DB.Model(&models.Service{}).Where("environment_id = ?", environmentID).Count(&count)
	return int(count), result.Error
}
