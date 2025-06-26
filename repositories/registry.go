package repositories

import (
	"fmt"
	"strings"

	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/models"
	"gorm.io/gorm"
)

// RegistryRepository handles database operations for registries
type RegistryRepository struct{}

// NewRegistryRepository creates a new registry repository instance
func NewRegistryRepository() *RegistryRepository {
	return &RegistryRepository{}
}

// FindAll retrieves all registries
func (r *RegistryRepository) FindAll() ([]models.Registry, error) {
	var registries []models.Registry
	result := database.DB.Find(&registries)
	return registries, result.Error
}

// FindByID retrieves a registry by its ID
func (r *RegistryRepository) FindByID(id string) (models.Registry, error) {
	var registry models.Registry
	result := database.DB.First(&registry, "id = ?", id)
	return registry, result.Error
}

// FindActive retrieves all active registries
func (r *RegistryRepository) FindActive() ([]models.Registry, error) {
	var registries []models.Registry
	result := database.DB.Where("is_active = ?", true).Find(&registries)
	return registries, result.Error
}

// FindDefault retrieves the default registry
func (r *RegistryRepository) FindDefault() (models.Registry, error) {
	var registry models.Registry
	result := database.DB.Where("is_default = ?", true).First(&registry)
	return registry, result.Error
}

// FindWithPagination retrieves registries with pagination, filtering and sorting
func (r *RegistryRepository) FindWithPagination(
	page, pageSize int,
	sortBy, sortOrder string,
	search string,
	onlyActive bool) ([]models.Registry, int64, error) {

	// Calculate offset
	offset := (page - 1) * pageSize

	// Valid sort columns (whitelist approach for security)
	validSortColumns := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"name":        true,
	}

	// Validate sort column
	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}

	// Validate sort order
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Build query
	query := database.DB.Model(&models.Registry{})

	// Add search filter if provided
	if search != "" {
		searchTerm := fmt.Sprintf("%%%s%%", search)
		query = query.Where("name ILIKE ? OR url ILIKE ?", searchTerm, searchTerm)
	}

	// Add active filter if requested
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	var registries []models.Registry
	result := query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Limit(pageSize).
		Offset(offset).
		Find(&registries)

	return registries, total, result.Error
}

// Create inserts a new registry into the database
func (r *RegistryRepository) Create(registry models.Registry) (models.Registry, error) {
	// If this registry is set as default, remove default flag from other registries
	if registry.IsDefault {
		database.DB.Model(&models.Registry{}).Where("is_default = ?", true).Update("is_default", false)
	}
	
	result := database.DB.Create(&registry)
	return registry, result.Error
}

// Update modifies an existing registry
func (r *RegistryRepository) Update(registry models.Registry) error {
	// If this registry is set as default, remove default flag from other registries
	if registry.IsDefault {
		database.DB.Model(&models.Registry{}).Where("id != ? AND is_default = ?", registry.ID, true).Update("is_default", false)
	}
	
	result := database.DB.Save(&registry)
	return result.Error
}

// Delete removes a registry from the database
func (r *RegistryRepository) Delete(id string) error {
	result := database.DB.Delete(&models.Registry{}, "id = ?", id)
	return result.Error
}

// SetActive updates the active status of a registry
func (r *RegistryRepository) SetActive(id string, active bool) error {
	result := database.DB.Model(&models.Registry{}).Where("id = ?", id).Update("is_active", active)
	return result.Error
}

// SetDefault sets a registry as the default and removes default flag from others
func (r *RegistryRepository) SetDefault(id string) error {
	// First, remove default flag from all registries
	if err := database.DB.Model(&models.Registry{}).Update("is_default", false).Error; err != nil {
		return err
	}
	
	// Then set the specified registry as default
	result := database.DB.Model(&models.Registry{}).Where("id = ?", id).Update("is_default", true)
	return result.Error
}

// DB returns the database instance
func (r *RegistryRepository) DB() *gorm.DB {
	return database.DB
}
