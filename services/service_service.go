package services

import (
	"errors"
	"fmt"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
)

// ServiceService handles business logic for services
type ServiceService struct {
	serviceRepo     *repositories.ServiceRepository
	projectRepo     *repositories.ProjectRepository
	environmentRepo *repositories.EnvironmentRepository
}

// NewServiceService creates a new service service instance
func NewServiceService() *ServiceService {
	return &ServiceService{
		serviceRepo:     repositories.NewServiceRepository(),
		projectRepo:     repositories.NewProjectRepository(),
		environmentRepo: repositories.NewEnvironmentRepository(),
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
	
	// Set initial status
	service.Status = "inactive"
	
	// Create the service
	return s.serviceRepo.Create(service)
}

// UpdateService updates an existing service
func (s *ServiceService) UpdateService(service models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Fetch the existing service
	existingService, err := s.serviceRepo.FindByID(service.ID)
	if err != nil {
		return service, fmt.Errorf("service not found: %v", err)
	}
	
	// Check if user can access this service's project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(existingService.ProjectID)
		if err != nil {
			return service, err
		}
		
		if ownerID != userID {
			return service, errors.New("unauthorized access to service")
		}
	}
	
	// Prevent changing project ID
	if service.ProjectID != existingService.ProjectID {
		return service, errors.New("cannot change project for an existing service")
	}
	
	// Verify environment exists and belongs to the project
	if service.EnvironmentID != existingService.EnvironmentID {
		env, err := s.environmentRepo.FindByID(service.EnvironmentID)
		if err != nil {
			return service, errors.New("environment not found")
		}
		
		if env.ProjectID != service.ProjectID {
			return service, errors.New("environment does not belong to the specified project")
		}
	}
	
	// Preserve fields that shouldn't be changed
	service.APIKey = existingService.APIKey
	service.Status = existingService.Status
	service.Domain = existingService.Domain
	service.CreatedAt = existingService.CreatedAt
	
	// Update the service
	err = s.serviceRepo.Update(service)
	if err != nil {
		return service, err
	}
	
	// Fetch the updated service with its relationships
	return s.serviceRepo.FindByID(service.ID)
}

// DeleteService deletes a service
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
	
	// Delete the service
	return s.serviceRepo.Delete(serviceID)
}
