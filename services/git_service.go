package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
	"github.com/pendeploy-simple/utils"
)

type GitService struct {
	projectRepo       *repositories.ProjectRepository
	environmentRepo   *repositories.EnvironmentRepository
	serviceRepo       *repositories.ServiceRepository
	deploymentRepo    *repositories.DeploymentRepository
	deploymentService *DeploymentService
}

// NewGitService creates a new git service instance
func NewGitService() *GitService {
	return &GitService{
		projectRepo:       repositories.NewProjectRepository(),
		environmentRepo:   repositories.NewEnvironmentRepository(),
		serviceRepo:       repositories.NewServiceRepository(),
		deploymentRepo:    repositories.NewDeploymentRepository(),
		deploymentService: NewDeploymentService(),
	}
}

// createGitService handles git service creation (MOVED from original CreateService)
func (s *GitService) CreateGitService(service models.Service, userID string, isAdmin bool) (models.Service, error) {
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

	// Validate service fields for git type
	if service.RepoURL == "" {
		return service, errors.New("repository URL is required for git services")
	}

	// Set default branch if empty
	if service.Branch == "" {
		service.Branch = "main"
	}

	// Set initial status
	service.Status = "inactive"

	go s.deploymentService.CreateGitDeployment(dto.GitDeployRequest{
		ServiceID:     service.ID,
		APIKey:        service.APIKey,
		CommitID:      "",
		CommitMessage: "Init",
	})

	// Create the service
	return s.serviceRepo.Create(service)
}

// updateGitService handles git service updates (MOVED from original UpdateService)
func (s *GitService) UpdateGitService(newService models.Service, userID string, isAdmin bool) (models.Service, error) {
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
		log.Println("update env vars")
		// If we want to completely replace env vars
		updatedService.EnvVars = newService.EnvVars
		log.Println(updatedService.EnvVars)
	}

	// Trigger redeployment for git services
	deployment, err := s.deploymentRepo.GetLatestDeployment(updatedService.ID)
	if err != nil {
		return newService, err
	}
	updatedService.Status = "building"
	// Update the service in the database
	err = s.serviceRepo.Update(updatedService)
	if err != nil {
		return newService, err
	}
	
	go s.deploymentService.CreateGitDeployment(dto.GitDeployRequest{
		ServiceID:     updatedService.ID,
		APIKey:        updatedService.APIKey,
		CommitID:      deployment.CommitSHA,
		CommitMessage: deployment.CommitMessage,
	})
	
	
	
	// Fetch the updated service with its relationships
	return s.serviceRepo.FindByID(newService.ID)
}

// deleteGitService handles git service deletion (MOVED from original DeleteService)
func (s *GitService) DeleteGitService(serviceID string, userID string, isAdmin bool) error {
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
	buildErr := utils.DeleteBuildResources(service)
	if buildErr != nil {
		// Log the error but continue with deletion
		fmt.Printf("Warning: Error deleting build resources for service %s: %v\n", serviceID, buildErr)
		return buildErr
	}
	
	// Step 3: Delete the service from database
	return s.serviceRepo.Delete(serviceID)
}