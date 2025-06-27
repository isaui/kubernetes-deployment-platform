package services

import (
	"errors"
	"fmt"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
	"github.com/pendeploy-simple/utils"
)

// ServiceService handles business logic for services
type ServiceService struct {
	serviceRepo     *repositories.ServiceRepository
	projectRepo     *repositories.ProjectRepository
	environmentRepo *repositories.EnvironmentRepository
	deploymentRepo  *repositories.DeploymentRepository
	deploymentService *DeploymentService
}

// NewServiceService creates a new service service instance
func NewServiceService() *ServiceService {
	return &ServiceService{
		serviceRepo:     repositories.NewServiceRepository(),
		projectRepo:     repositories.NewProjectRepository(),
		environmentRepo: repositories.NewEnvironmentRepository(),
		deploymentRepo:  repositories.NewDeploymentRepository(),
		deploymentService:  NewDeploymentService(),
	}
}

// ListAllServices retrieves all services (admin only)
func (s *ServiceService) ListAllServices() ([]models.Service, error) {
	return s.serviceRepo.FindAll()
}

// ListProjectServices retrieves all services for a project
func (s *ServiceService) ListProjectServices(projectID string, userID string, isAdmin bool) ([]models.Service, error) {
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(projectID)
		if err != nil {
			return nil, err
		}
		
		if ownerID != userID {
			return nil, errors.New("unauthorized access to project services")
		}
	}
	
	return s.serviceRepo.FindByProjectID(projectID)
}

// GetServiceDetail retrieves a specific service
func (s *ServiceService) GetServiceDetail(serviceID string, userID string, isAdmin bool) (models.Service, error) {
	// Fetch the service with deployments
	service, err := s.serviceRepo.WithDeployments(serviceID)
	if err != nil {
		return service, err
	}
	
	// Check if user can access this service's project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(service.ProjectID)
		if err != nil {
			return service, err
		}
		
		if ownerID != userID {
			return service, errors.New("unauthorized access to service")
		}
	}
	
	return service, nil
}

// CreateService creates a new service
func (s *ServiceService) CreateService(service models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(service.ProjectID)
		if err != nil {
			return service, err
		}
		
		if ownerID != userID {
			return service, errors.New("unauthorized access to project")
		}
	}
	
	// Verify environment exists and belongs to the project
	env, err := s.environmentRepo.FindByID(service.EnvironmentID)
	if err != nil {
		return service, errors.New("environment not found")
	}
	
	if env.ProjectID != service.ProjectID {
		return service, errors.New("environment does not belong to the specified project")
	}
	
	// Validate service fields based on type
	if service.Type == models.ServiceTypeGit {
		// Git-based service requires RepoURL
		if service.RepoURL == "" {
			return service, errors.New("repository URL is required for git services")
		}
		
		// Set default branch if empty
		if service.Branch == "" {
			service.Branch = "main"
		}
	} else if service.Type == models.ServiceTypeManaged {
		// Managed service requires ManagedType
		if service.ManagedType == "" {
			return service, errors.New("managed type is required for managed services")
		}
		
		// Set default version if empty
		if service.Version == "" {
			service.Version = "latest"
		}
		
		// Set default storage size for database-like managed services
		if service.StorageSize == "" && (service.ManagedType == "postgresql" || 
			service.ManagedType == "mysql" || service.ManagedType == "mongodb" || 
			service.ManagedType == "redis" || service.ManagedType == "minio") {
			service.StorageSize = "1Gi"
		}
	} else {
		return service, errors.New("invalid service type")
	}

	// Set initial status
	service.Status = "inactive"
	
	// Create the service
	return s.serviceRepo.Create(service)
}

// UpdateService updates an existing service using selective update approach
// This ensures that only provided fields are updated and other fields keep their existing values
func (s *ServiceService) UpdateService(newService models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Fetch the existing service
	existingService, err := s.serviceRepo.FindByID(newService.ID)
	if err != nil {
		return newService, fmt.Errorf("service not found: %v", err)
	}
	
	// Check if user can access this service's project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(existingService.ProjectID)
		if err != nil {
			return newService, err
		}
		
		if ownerID != userID {
			return newService, errors.New("unauthorized access to service")
		}
	}
	
	// Start with existing service - this is important for selective update
	updatedService := existingService
	
	// Only update fields that should be allowed to change
	// Always allow name update
	if newService.Name != "" {
		updatedService.Name = newService.Name
	}
	
	// Prevent changing project ID
	if newService.ProjectID != "" && newService.ProjectID != existingService.ProjectID {
		return newService, errors.New("cannot change project for an existing service")
	}
	
	// If environment ID is provided, verify it exists and belongs to the project
	if newService.EnvironmentID != "" && newService.EnvironmentID != existingService.EnvironmentID {
		env, err := s.environmentRepo.FindByID(newService.EnvironmentID)
		if err != nil {
			return newService, errors.New("environment not found")
		}
		
		if env.ProjectID != existingService.ProjectID {
			return newService, errors.New("environment does not belong to the specified project")
		}
		
		updatedService.EnvironmentID = newService.EnvironmentID
	}
	
	// Type-specific updates based on existing service type
	if existingService.Type == models.ServiceTypeGit {
		// For Git services, allow updating branch and port
		if newService.Branch != "" {
			updatedService.Branch = newService.Branch
		}
		
		// Port can be 0 (use default) or a positive value
		if newService.Port > 0 {
			updatedService.Port = newService.Port
		}
		
		// Allow updating build/start commands if provided
		if newService.BuildCommand != "" {
			updatedService.BuildCommand = newService.BuildCommand
		}
		
		if newService.StartCommand != "" {
			updatedService.StartCommand = newService.StartCommand
		}
		
		// RepoURL should not change - it's fundamental to the service
		// updatedService.RepoURL = existingService.RepoURL (already set because we started with existingService)
		
	} else if existingService.Type == models.ServiceTypeManaged {
		// For Managed services, allow updating version and storage size
		if newService.Version != "" {
			updatedService.Version = newService.Version
		}
		
		if newService.StorageSize != "" {
			updatedService.StorageSize = newService.StorageSize
		}
		
		// ManagedType should not change for existing services
		if newService.ManagedType != "" && newService.ManagedType != existingService.ManagedType {
			return newService, errors.New("cannot change managed service type for an existing service")
		}
	}
	
	// Update resource constraints if provided
	if newService.CPULimit != "" {
		updatedService.CPULimit = newService.CPULimit
	}
	
	if newService.MemoryLimit != "" {
		updatedService.MemoryLimit = newService.MemoryLimit
	}
	
	// Update replica configuration if provided
	if newService.IsStaticReplica != existingService.IsStaticReplica {
		updatedService.IsStaticReplica = newService.IsStaticReplica
	}
	
	if newService.Replicas > 0 {
		updatedService.Replicas = newService.Replicas
	}
	
	if newService.MinReplicas > 0 {
		updatedService.MinReplicas = newService.MinReplicas
	}
	
	if newService.MaxReplicas > 0 {
		updatedService.MaxReplicas = newService.MaxReplicas
	}
	
	// Update custom domain if provided
	if newService.CustomDomain != "" {
		updatedService.CustomDomain = newService.CustomDomain
	}
	
	// Update environment variables if provided
	if newService.EnvVars != nil && len(newService.EnvVars) > 0 {
		// If we want to completely replace env vars
		updatedService.EnvVars = newService.EnvVars
	}

	if(newService.Type == models.ServiceTypeGit){
		deployment, err := s.deploymentRepo.GetLatestSuccessfulDeployment(updatedService.ID)
		if err != nil {
			return newService, err
		}
		
		go s.deploymentService.CreateGitDeployment(dto.GitDeployRequest{
			ServiceID: updatedService.ID,
			APIKey: updatedService.APIKey,
			CommitID: deployment.CommitSHA,
			CommitMessage: deployment.CommitMessage,
		})
	}
	
	// Update the service in the database
	err = s.serviceRepo.Update(updatedService)
	if err != nil {
		return newService, err
	}
	
	// Fetch the updated service with its relationships
	return s.serviceRepo.FindByID(newService.ID)
}

// DeleteService deletes a service and all associated Kubernetes resources
func (s *ServiceService) DeleteService(serviceID string, userID string, isAdmin bool) error {
	// Fetch the service
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Check if user can access this service's project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(service.ProjectID)
		if err != nil {
			return err
		}
		
		if ownerID != userID {
			return errors.New("unauthorized access to service")
		}
	}

	// Step 1: Delete Kubernetes resources first
	err = utils.DeleteKubernetesResources(service)
	if err != nil {
		// Log the error but continue with database deletion
		fmt.Printf("Warning: Error deleting Kubernetes resources for service %s: %v\n", serviceID, err)
		return err
	}
	
	// Step 2: Clean up any build jobs related to this service if they exist
	// The build jobs are in 'build-and-deploy' namespace with service ID in their name
	buildErr := utils.DeleteBuildResources(service)
	if buildErr != nil {
		// Log the error but continue with deletion
		fmt.Printf("Warning: Error deleting build resources for service %s: %v\n", serviceID, buildErr)
		return buildErr
	}
	
	// Step 3: Delete the service from database
	return s.serviceRepo.Delete(serviceID)
}
