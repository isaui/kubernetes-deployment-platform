package repositories

import (
	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/models"
	"gorm.io/gorm"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct{}

// NewProjectRepository creates a new project repository instance
func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}

// FindAll retrieves all projects
func (r *ProjectRepository) FindAll() ([]models.Project, error) {
	var projects []models.Project
	result := database.DB.Find(&projects)
	return projects, result.Error
}

// FindByID retrieves a project by its ID
func (r *ProjectRepository) FindByID(id string) (models.Project, error) {
	var project models.Project
	result := database.DB.First(&project, "id = ?", id)
	return project, result.Error
}

// FindByUserID retrieves all projects belonging to a user
func (r *ProjectRepository) FindByUserID(userID string) ([]models.Project, error) {
	var projects []models.Project
	result := database.DB.Where("user_id = ?", userID).Find(&projects)
	return projects, result.Error
}

// Create inserts a new project into the database
func (r *ProjectRepository) Create(project models.Project) (models.Project, error) {
	result := database.DB.Create(&project)
	return project, result.Error
}

// Update modifies an existing project
func (r *ProjectRepository) Update(project models.Project) error {
	result := database.DB.Save(&project)
	return result.Error
}

// Delete removes a project from the database (soft delete)
func (r *ProjectRepository) Delete(id string) error {
	// Gunakan transaction untuk memastikan konsistensi data
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Soft delete service terkait terlebih dahulu
		if err := tx.Where("project_id = ?", id).Delete(&models.Service{}).Error; err != nil {
			return err
		}
		
		// Soft delete project
		result := tx.Delete(&models.Project{}, "id = ?", id)
		return result.Error
	})
}

// Exists checks if a project exists (including soft-deleted ones)
func (r *ProjectRepository) Exists(id string) (bool, error) {
	var count int64
	// Menggunakan Unscoped() untuk melihat semua record termasuk yang soft-deleted
	err := database.DB.Unscoped().Model(&models.Project{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// GetOwnerID returns the user ID who owns the project
func (r *ProjectRepository) GetOwnerID(id string) (string, error) {
	type ProjectOwner struct {
		UserID string
	}
	
	var owner ProjectOwner
	// Menggunakan Unscoped() untuk mendapatkan project meskipun sudah soft-deleted
	err := database.DB.Unscoped().Model(&models.Project{}).Select("user_id").Where("id = ?", id).First(&owner).Error
	return owner.UserID, err
}

// CountByUserID counts projects belonging to a user
func (r *ProjectRepository) CountByUserID(userID string) (int64, error) {
	var count int64
	result := database.DB.Model(&models.Project{}).Where("user_id = ?", userID).Count(&count)
	return count, result.Error
}

// DB returns the database instance
func (r *ProjectRepository) DB() *gorm.DB {
	return database.DB
}

// WithServices loads project with its services
func (r *ProjectRepository) WithServices(id string) (models.Project, error) {
	var project models.Project
	result := database.DB.Preload("Services").First(&project, "id = ?", id)
	return project, result.Error
}

// FindWithPagination retrieves projects with pagination, filtering and sorting
func (r *ProjectRepository) FindWithPagination(
	page, pageSize int,
	sortBy, sortOrder string,
	userID string,
	isAdmin bool,
	search string) ([]models.Project, int64, error) {

	var projects []models.Project
	var totalCount int64
	
	// Buat query dasar
	db := database.DB.Model(&models.Project{})
	
	// Filter by user ID jika bukan admin
	if !isAdmin && userID != "" {
		db = db.Where("user_id = ?", userID)
	}
	
	// Search functionality
	if search != "" {
		searchPattern := "%" + search + "%"
		db = db.Where("(name ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
	}
	
	// Count total records (dengan filter yang sama)
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	
	// Calculate offset for pagination
	offset := (page - 1) * pageSize
	
	// Sort dan paginate
	orderString := sortBy + " " + sortOrder
	if err := db.Order(orderString).Limit(pageSize).Offset(offset).Find(&projects).Error; err != nil {
		return nil, 0, err
	}
	
	return projects, totalCount, nil
}
