package services

import (
	"errors"
	"fmt"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
)

// ServiceService handles business logic for services (UPDATED untuk managed services)
type ServiceService struct {
	serviceRepo       *repositories.ServiceRepository
	projectRepo       *repositories.ProjectRepository
	environmentRepo   *repositories.EnvironmentRepository
	deploymentRepo    *repositories.DeploymentRepository
	deploymentService *DeploymentService
	gitService        *GitService
	managedService    *ManagedServiceService // NEW: Managed service handler
}

// NewServiceService creates a new service service instance (UPDATED)
func NewServiceService() *ServiceService {
	return &ServiceService{
		serviceRepo:       repositories.NewServiceRepository(),
		projectRepo:       repositories.NewProjectRepository(),
		environmentRepo:   repositories.NewEnvironmentRepository(),
		deploymentRepo:    repositories.NewDeploymentRepository(),
		deploymentService: NewDeploymentService(),
		gitService:        NewGitService(),
		managedService:    NewManagedServiceService(), // NEW
	}
}

func (s *ServiceService) GetDeploymentList(serviceID string, userID string, isAdmin bool) ([]dto.DeploymentResponse, error) {
	deployments, err := s.deploymentRepo.FindByServiceID(serviceID)
	if err != nil {
		return nil, err
	}
	
	// Map deployments to DTOs for API stability
	deploymentResponses := make([]dto.DeploymentResponse, len(deployments))
	for i, deployment := range deployments {
		deploymentResponses[i] = dto.NewDeploymentResponseFromModel(deployment)
	}
	
	return deploymentResponses, nil
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

// CreateService creates a new service - UPDATED untuk handle managed services
func (s *ServiceService) CreateService(service models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Route to appropriate service type handler
	switch service.Type {
	case models.ServiceTypeGit:
		return s.gitService.CreateGitService(service, userID, isAdmin)
	case models.ServiceTypeManaged:
		return s.managedService.CreateManagedService(service, userID, isAdmin) // NEW
	default:
		return service, errors.New("invalid service type")
	}
}

// UpdateService updates an existing service - UPDATED untuk handle managed services
func (s *ServiceService) UpdateService(newService models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Get existing service to determine type
	existingService, err := s.serviceRepo.FindByID(newService.ID)
	if err != nil {
		return newService, fmt.Errorf("service not found: %v", err)
	}
	
	// Route to appropriate service type handler
	switch existingService.Type {
	case models.ServiceTypeGit:
		return s.gitService.UpdateGitService(newService, userID, isAdmin)
	case models.ServiceTypeManaged:
		return s.managedService.UpdateManagedService(newService, userID, isAdmin) // NEW
	default:
		return newService, errors.New("invalid service type")
	}
}


/// DeleteService deletes a service - UPDATED untuk handle managed services
func (s *ServiceService) DeleteService(serviceID string, userID string, isAdmin bool) error {
	// Get service to determine type
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Route to appropriate service type handler
	switch service.Type {
	case models.ServiceTypeGit:
		return s.gitService.DeleteGitService(serviceID, userID, isAdmin)
	case models.ServiceTypeManaged:
		return s.managedService.DeleteManagedService(serviceID, userID, isAdmin) // NEW
	default:
		return errors.New("invalid service type")
	}
}

func (s *ServiceService) GetLatestDeployment(serviceID string, userID string, isAdmin bool) (dto.DeploymentResponse, error) {
	// Verify this is a git service first
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return dto.DeploymentResponse{}, err
	}
	
	if service.Type != models.ServiceTypeGit {
		return dto.DeploymentResponse{}, errors.New("deployments are only available for git services")
	}
	
	deployment, deployErr := s.deploymentRepo.GetLatestDeployment(serviceID)
	if deployErr != nil {
		return dto.DeploymentResponse{}, deployErr
	}
	
	// Check if user can access this deployment's project
	if !isAdmin {
		ownerID, ownerErr := s.projectRepo.GetOwnerID(service.ProjectID)
		if ownerErr != nil {
			return dto.DeploymentResponse{}, ownerErr
		}
		
		if ownerID != userID {
			return dto.DeploymentResponse{}, errors.New("unauthorized access to deployment")
		}
	}
	
	return dto.NewDeploymentResponseFromModel(deployment), nil
}


