package services

import (
	"fmt"
	"log"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	projectRepo *repositories.ProjectRepository
	environmentRepo *repositories.EnvironmentRepository
}

// NewProjectService creates a new project service instance
func NewProjectService() *ProjectService {
	return &ProjectService{
		projectRepo: repositories.NewProjectRepository(),
		environmentRepo: repositories.NewEnvironmentRepository(),
	}
}



// ListProjects retrieves projects with pagination, filtering and sorting
// Admin can see all projects, regular users only see their own
func (s *ProjectService) ListProjects(filter dto.ProjectFilter) (dto.ProjectListResponse, error) {
	var response dto.ProjectListResponse
	
	// Set defaults if not provided
	if filter.Page <= 0 {
		filter.Page = 1
	}
	
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}
	
	// Validate sort order
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "desc"
	}
	
	// Valid sort columns (whitelist approach for security)
	validSortColumns := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"name":        true,
	}
	
	if !validSortColumns[filter.SortBy] {
		filter.SortBy = "created_at"
	}
	
	// Gunakan repository untuk mengambil data (dengan pagination, filtering, sorting)
	projects, totalCount, err := s.projectRepo.FindWithPagination(
		filter.Page,
		filter.PageSize,
		filter.SortBy,
		filter.SortOrder,
		filter.UserID,
		filter.IsAdmin,
		filter.Search,
	)
	
	if err != nil {
		return response, err
	}
	
	// Calculate total pages
	totalPages := int(totalCount) / filter.PageSize
	if int(totalCount)%filter.PageSize > 0 {
		totalPages++
	}
	
	// Build response
	response = dto.ProjectListResponse{
		Projects:   projects,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}
	
	return response, nil
}

// GetProjectDetail retrieves a project by ID with its services
// Access control: admin can view any project, regular users only their own
func (s *ProjectService) GetProjectDetail(projectID string, userID string, isAdmin bool) (models.Project, error) {
	// Retrieve project with services
	project, err := s.projectRepo.WithServices(projectID)
	if err != nil {
		return models.Project{}, err
	}
	
	// Access control - return error if not admin and not owner
	if !isAdmin && project.UserID != userID {
		return models.Project{}, fmt.Errorf("unauthorized: you don't have permission to access this project")
	}
	
	return project, nil
}

// GetProjectStats retrieves statistics for a project
func (s *ProjectService) GetProjectStats(projectID string, userID string, isAdmin bool) (dto.ProjectStatsResponse, error) {
	// First check access permissions
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return dto.ProjectStatsResponse{}, err
	}
	
	// Access control - return error if not admin and not owner
	if !isAdmin && project.UserID != userID {
		return dto.ProjectStatsResponse{}, fmt.Errorf("unauthorized: you don't have permission to access this project")
	}
	
	// Get project environments
	environmentRepo := repositories.NewEnvironmentRepository()
	environments, err := environmentRepo.FindByProjectID(projectID)
	if err != nil {
		return dto.ProjectStatsResponse{}, err
	}
	
	// Get project services
	serviceRepo := repositories.NewServiceRepository()
	services, err := serviceRepo.FindByProjectID(projectID)
	if err != nil {
		return dto.ProjectStatsResponse{}, err
	}
	
	// Get deployment stats
	deploymentRepo := repositories.NewDeploymentRepository()
	
	// Prepare stats result
	stats := dto.ProjectStatsResponse{}
	
	// Set project info
	stats.Project.ID = project.ID
	stats.Project.Name = project.Name
	stats.Project.Description = project.Description
	stats.Project.CreatedAt = project.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	
	// Initialize environment stats
	stats.Environments.Total = len(environments)
	stats.Environments.Environments = make([]dto.ProjectEnvironmentItem, 0)
	
	// Process environments
	envMap := make(map[string]models.Environment)
	for _, env := range environments {
		envMap[env.ID] = env
		
		// Count services in this environment
		serviceCount, _ := environmentRepo.CountServicesInEnvironment(env.ID)
		
		// Add environment to list
		envItem := dto.ProjectEnvironmentItem{
			ID:            env.ID,
			Name:          env.Name,
			Description:   env.Description,
			ServicesCount: serviceCount,
			CreatedAt:     env.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		stats.Environments.Environments = append(stats.Environments.Environments, envItem)
	}
	
	// Initialize service stats
	stats.Services.Total = len(services)
	stats.Services.ByType = make(map[string]int)
	stats.Services.ByStatus = make(map[string]int)
	stats.Services.ServiceList = make([]dto.ProjectServiceStatsItem, 0)
	
	// Initialize deployment stats
	stats.Deployments.Total = 0
	stats.Deployments.Successful = 0
	stats.Deployments.Failed = 0
	stats.Deployments.InProgress = 0
	stats.Deployments.SuccessRate = 0.0
	
	// Variables to track total deployments
	var totalDeployments int64 = 0
	
	// Get deployments stats directly from repository using the project ID
	// This is more efficient than counting per service
	totalDeployments, _ = deploymentRepo.CountByProjectID(projectID)
	successfulDeployments, _ := deploymentRepo.CountDeploymentsByProjectIDAndStatus(projectID, models.DeploymentStatusSuccess)
	failedDeployments, _ := deploymentRepo.CountDeploymentsByProjectIDAndStatus(projectID, models.DeploymentStatusFailed)
	inProgressDeployments, _ := deploymentRepo.CountDeploymentsByProjectIDAndStatus(projectID, models.DeploymentStatusBuilding)
	
	// Process each service
	for _, service := range services {
		// Increment service type counter
		stats.Services.ByType[string(service.Type)]++
		
		// Increment status counter
		stats.Services.ByStatus[service.Status]++
		
		// Get deployments for this service
		deploymentCount, _ := deploymentRepo.CountByServiceID(service.ID)
		
		// Get successful deployments
		successRate, _ := deploymentRepo.GetSuccessRate(service.ID)
		
		// Get environment information
		var environmentID string = service.EnvironmentID
		var environmentName string = ""
		
		// Look up environment name from our map
		if env, exists := envMap[environmentID]; exists {
			environmentName = env.Name
		}
		
		// Add service to list
		serviceItem := dto.ProjectServiceStatsItem{
			ID:              service.ID,
			Name:            service.Name,
			Type:            string(service.Type),
			Status:          service.Status,
			EnvironmentID:   environmentID,
			EnvironmentName: environmentName,
			Deployments:     deploymentCount,
			SuccessRate:     successRate,
			Replicas:        service.Replicas,
			IsAutoScaling:   !service.IsStaticReplica,
		}
		stats.Services.ServiceList = append(stats.Services.ServiceList, serviceItem)
	}
	
	// Update deployment stats
	stats.Deployments.Total = totalDeployments
	stats.Deployments.Successful = successfulDeployments
	stats.Deployments.Failed = failedDeployments
	stats.Deployments.InProgress = inProgressDeployments
	
	if totalDeployments > 0 {
		stats.Deployments.SuccessRate = float64(successfulDeployments) / float64(totalDeployments)
	}
	
	return stats, nil
}

// CreateProject creates a new project with a default environment
func (s *ProjectService) CreateProject(project models.Project) (models.Project, error) {
	// Begin a transaction to ensure both project and environment are created together
	db := s.projectRepo.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()

	// Create project
	if err := db.Create(&project).Error; err != nil {
		db.Rollback()
		return project, err
	}

	// Create default environment
	defaultEnv := models.Environment{
		Name:        "Production",
		Description: "Default production environment",
		ProjectID:   project.ID,
	}

	if err := db.Create(&defaultEnv).Error; err != nil {
		db.Rollback()
		return project, err
	}

	// Commit the transaction
	if err := db.Commit().Error; err != nil {
		return project, err
	}

	return project, nil
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(project models.Project, userID string, isAdmin bool) (models.Project, error) {
	// Get existing project
	existingProject, err := s.projectRepo.FindByID(project.ID)
	if err != nil {
		return models.Project{}, err
	}
	
	// Access control - return error if not admin and not owner
	if !isAdmin && existingProject.UserID != userID {
		return models.Project{}, fmt.Errorf("unauthorized: you don't have permission to update this project")
	}
	
	// Preserve the user ID (it shouldn't be changed)
	project.UserID = existingProject.UserID
	
	// Update project
	err = s.projectRepo.Update(project)
	if err != nil {
		return models.Project{}, err
	}
	
	return project, nil
}

// DeleteProject deletes a project and all associated kubernetes resources
func (s *ProjectService) DeleteProject(projectID string, userID string, isAdmin bool) error {
	// Cek apakah project dengan ID tersebut ada di database (tanpa filter deleted_at)
	exists, err := s.projectRepo.Exists(projectID)
	if err != nil {
		return err
	}
	
	if !exists {
		return fmt.Errorf("project not found or already deleted")
	}
	
	// Jika kita ingin memvalidasi akses kontrol, maka kita perlu mendapatkan project
	// dengan Unscoped (untuk melihat bahkan yang sudah dihapus)
	if !isAdmin {
		owner, err := s.projectRepo.GetOwnerID(projectID)
		if err != nil {
			return err
		}
			
		if owner != userID {
			return fmt.Errorf("unauthorized: you don't have permission to delete this project")
		}
	}

	// Get all project environments before deleting
	environments, err := s.environmentRepo.FindByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("error fetching project environments: %w", err)
	}

	// Init Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("error initializing kubernetes client: %w", err)
	}

	// Delete all kubernetes namespaces for each environment
	for _, env := range environments {
		// Delete namespace for environment
		namespace := env.ID // The namespace name is the environment ID
		
		// Check if namespace exists
		exists, err := k8sClient.NamespaceExists(namespace)
		if err != nil {
			log.Printf("Warning: Error checking namespace %s: %v", namespace, err)
			continue
		}

		if exists {
			err = k8sClient.DeleteNamespace(namespace)
			if err != nil {
				log.Printf("Warning: Failed to delete namespace %s: %v", namespace, err)
				// Continue with other namespaces even if one fails
			} else {
				log.Printf("Successfully deleted namespace %s for environment %s", namespace, env.Name)
			}
		}
	}

	// Lakukan soft delete - cascade will handle related records
	return s.projectRepo.Delete(projectID)
}
