// services/managed_service.go
package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
	"github.com/pendeploy-simple/utils"
	
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// ManagedServiceService handles business logic for managed services
type ManagedServiceService struct {
	serviceRepo     *repositories.ServiceRepository
	projectRepo     *repositories.ProjectRepository
	environmentRepo *repositories.EnvironmentRepository
}

// NewManagedServiceService creates a new managed service service instance
func NewManagedServiceService() *ManagedServiceService {
	return &ManagedServiceService{
		serviceRepo:     repositories.NewServiceRepository(),
		projectRepo:     repositories.NewProjectRepository(),
		environmentRepo: repositories.NewEnvironmentRepository(),
	}
}

// CreateManagedService creates and deploys a new managed service
func (s *ManagedServiceService) CreateManagedService(service models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Validate user access to project
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
	
	// Validate managed service configuration
	if err := s.validateManagedServiceConfig(service); err != nil {
		return service, err
	}
	
	// Set defaults for managed service
	service = s.setManagedServiceDefaults(service)
	service.Status = "building"
	
	// Create service in database first
	createdService, err := s.serviceRepo.Create(service)
	if err != nil {
		return service, fmt.Errorf("failed to create service in database: %v", err)
	}

	go func ()  {
		// Deploy to Kubernetes
		deployedService, err := s.deployManagedServiceToKubernetes(createdService)
		if err != nil {
			log.Printf("Failed to deploy managed service %s: %v", createdService.ID, err)
			
			// Update service status to failed
			deployedService.Status = "failed"
			s.serviceRepo.Update(*deployedService)
			
			log.Printf("Failed to deploy managed service %s: %v", createdService.ID, err)
		}
	    deployedService.Status = "active"
		// Update service with deployment results (env vars, domain, etc.)
		err = s.serviceRepo.Update(*deployedService)
		if err != nil {
			log.Printf("Failed to update service after deployment: %v", err)
			// Don't fail the entire operation, just log the error
		}
	}()
	
	log.Printf("Successfully created managed service: %s (%s)", createdService.Name, createdService.ManagedType)
	return createdService, nil
}

// UpdateManagedService updates an existing managed service
func (s *ManagedServiceService) UpdateManagedService(serviceChanges models.Service, userID string, isAdmin bool) (models.Service, error) {
	// Fetch the existing service
	existingService, err := s.serviceRepo.FindByID(serviceChanges.ID)
	if err != nil {
		return serviceChanges, fmt.Errorf("service not found: %v", err)
	}
	
	// Validate it's a managed service
	if existingService.Type != models.ServiceTypeManaged {
		return serviceChanges, errors.New("service is not a managed service")
	}
	
	// Check user access
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(existingService.ProjectID)
		if err != nil {
			return serviceChanges, err
		}
		
		if ownerID != userID {
			return serviceChanges, errors.New("unauthorized access to service")
		}
	}
	
	// Start with existing service for selective updates
	updatedService := existingService
	
	// Allow name updates
	if serviceChanges.Name != "" {
		updatedService.Name = serviceChanges.Name
	}
	
	// Prevent changing core fields
	if serviceChanges.ProjectID != "" && serviceChanges.ProjectID != existingService.ProjectID {
		return serviceChanges, errors.New("cannot change project for an existing service")
	}
	
	if serviceChanges.ManagedType != "" && serviceChanges.ManagedType != existingService.ManagedType {
		return serviceChanges, errors.New("cannot change managed service type")
	}
	
	// Allow environment changes (with validation)
	if serviceChanges.EnvironmentID != "" && serviceChanges.EnvironmentID != existingService.EnvironmentID {
		env, err := s.environmentRepo.FindByID(serviceChanges.EnvironmentID)
		if err != nil {
			return serviceChanges, errors.New("environment not found")
		}
		
		if env.ProjectID != existingService.ProjectID {
			return serviceChanges, errors.New("environment does not belong to the specified project")
		}
		
		updatedService.EnvironmentID = serviceChanges.EnvironmentID
	}
	
	// Allow resource updates
	if serviceChanges.CPULimit != "" {
		updatedService.CPULimit = serviceChanges.CPULimit
	}
	
	if serviceChanges.MemoryLimit != "" {
		updatedService.MemoryLimit = serviceChanges.MemoryLimit
	}
	
	// Allow storage size updates (only allow increase, StatefulSet cannot shrink storage)
	if serviceChanges.StorageSize != "" {
		if err := s.validateStorageSizeIncrease(existingService.StorageSize, serviceChanges.StorageSize); err != nil {
			return serviceChanges, err
		}
		updatedService.StorageSize = serviceChanges.StorageSize
	}
	
	// Allow version updates (will trigger redeployment)
	if serviceChanges.Version != "" {
		updatedService.Version = serviceChanges.Version
	}
	
	// Allow custom domain updates
	if serviceChanges.CustomDomain != "" {
		updatedService.CustomDomain = serviceChanges.CustomDomain
	}
	
	// Note: EnvVars are auto-generated and read-only for managed services
	// We don't allow user modifications
	
	// Set update timestamp
	updatedService.UpdatedAt = time.Now()
	
	// Redeploy to Kubernetes if significant changes were made
	needsRedeployment := s.checkIfRedeploymentNeeded(existingService, updatedService)
	
	if needsRedeployment {
		log.Printf("Redeploying managed service %s due to configuration changes", updatedService.ID)
		updatedService.Status = "building"
		go func () {
			// Redeploy to Kubernetes
			redeployedService, err := s.deployManagedServiceToKubernetes(updatedService)
			if err != nil {
				log.Printf("Failed to redeploy managed service %s: %v", updatedService.ID, err)
				
				// Update status to failed but keep other changes
				updatedService.Status = "failed"
				s.serviceRepo.Update(updatedService)
				
				log.Printf("Failed to update service after redeploy: %v", err)
				return
			}
			
			// Update with successful redeployment status and save to database
			updatedService = *redeployedService
			err = s.serviceRepo.Update(updatedService)
			if err != nil {
				log.Printf("Failed to update service status after successful redeploy: %v", err)
			}
			log.Printf("Successfully updated service status to: %s", updatedService.Status)
		}()
	}
	
	// Update the service in the database
	err = s.serviceRepo.Update(updatedService)
	if err != nil {
		return updatedService, fmt.Errorf("failed to update service in database: %v", err)
	}
	
	log.Printf("Successfully updated managed service: %s", updatedService.Name)
	return updatedService, nil
}

// DeleteManagedService deletes a managed service and cleans up Kubernetes resources
func (s *ManagedServiceService) DeleteManagedService(serviceID string, userID string, isAdmin bool) error {
	// Fetch the service
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Validate it's a managed service
	if service.Type != models.ServiceTypeManaged {
		return errors.New("service is not a managed service")
	}
	
	// Check user access
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(service.ProjectID)
		if err != nil {
			return err
		}
		
		if ownerID != userID {
			return errors.New("unauthorized access to service")
		}
	}
	
	// Delete Kubernetes resources
	err = s.deleteManagedServiceFromKubernetes(service)
	if err != nil {
		log.Printf("Warning: Error deleting Kubernetes resources for managed service %s: %v", serviceID, err)
		// Continue with database deletion even if K8s cleanup fails
	}
	
	// Delete from database
	err = s.serviceRepo.Delete(serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete service from database: %v", err)
	}
	
	log.Printf("Successfully deleted managed service: %s", serviceID)
	return nil
}

// validateManagedServiceConfig validates managed service configuration
func (s *ManagedServiceService) validateManagedServiceConfig(service models.Service) error {
	// Validate service type
	if service.Type != models.ServiceTypeManaged {
		return errors.New("service type must be 'managed'")
	}
	
	// Validate managed type
	if !utils.IsValidManagedServiceType(service.ManagedType) {
		return fmt.Errorf("unsupported managed service type: %s", service.ManagedType)
	}
	
	// Validate storage size format if provided
	if service.StorageSize != "" {
		// This is a basic validation - Kubernetes will do more thorough validation
		if len(service.StorageSize) < 2 {
			return errors.New("invalid storage size format")
		}
	}
	
	// Validate CPU and memory limits if provided
	if service.CPULimit != "" {
		// Basic validation - should end with 'm' or be a number
		if len(service.CPULimit) < 2 {
			return errors.New("invalid CPU limit format")
		}
	}
	
	if service.MemoryLimit != "" {
		// Basic validation - should end with 'Mi', 'Gi', etc.
		if len(service.MemoryLimit) < 3 {
			return errors.New("invalid memory limit format")
		}
	}
	
	return nil
}

// validateStorageSizeIncrease validates that storage size can only be increased
func (s *ManagedServiceService) validateStorageSizeIncrease(existingSize, newSize string) error {
	if existingSize == "" {
		return nil // No existing size, allow any size
	}
	
	// Use Kubernetes resource parsing
	existingQuantity, err := resource.ParseQuantity(existingSize)
	if err != nil {
		return fmt.Errorf("invalid existing storage size: %v", err)
	}
	
	newQuantity, err := resource.ParseQuantity(newSize)
	if err != nil {
		return fmt.Errorf("invalid new storage size: %v", err)
	}
	
	// Compare quantities
	if newQuantity.Cmp(existingQuantity) < 0 {
		return fmt.Errorf("storage size cannot be decreased from %s to %s", existingSize, newSize)
	}
	
	return nil
}

// setManagedServiceDefaults sets default values for managed service
func (s *ManagedServiceService) setManagedServiceDefaults(service models.Service) models.Service {
	// Set default version if empty
	if service.Version == "" {
		service.Version = utils.GetManagedServiceDefaultVersion(service.ManagedType)
	}
	
	// Set default storage size if empty and storage is required
	if service.StorageSize == "" && utils.RequiresPersistentStorage(service.ManagedType) {
		service.StorageSize = "1Gi"
	}
	
	// Set default resource limits if empty
	if service.CPULimit == "" {
		service.CPULimit = "500m"
	}
	
	if service.MemoryLimit == "" {
		service.MemoryLimit = "512Mi"
	}
	
	// Managed services are always single replica for data consistency
	service.IsStaticReplica = true
	service.Replicas = 1
	service.MinReplicas = 1
	service.MaxReplicas = 1
	
	// Set initial status
	service.Status = "inactive"
	
	// Set timestamps
	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now
	
	return service
}

// deployManagedServiceToKubernetes deploys the managed service to Kubernetes
func (s *ManagedServiceService) deployManagedServiceToKubernetes(service models.Service) (*models.Service, error) {
	log.Printf("Deploying managed service %s (%s) to Kubernetes", service.Name, service.ManagedType)
	serverIP := os.Getenv("SERVER_IP")
	
	// Use the Kubernetes deployment utility
	deployedService, err := utils.DeployManagedServiceToKubernetes(service, serverIP)
	if err != nil {
		return deployedService, err
	}
	
	log.Printf("Managed service %s deployed successfully", service.Name)
	return deployedService, nil
}

// deleteManagedServiceFromKubernetes deletes managed service resources from Kubernetes
func (s *ManagedServiceService) deleteManagedServiceFromKubernetes(service models.Service) error {
	log.Printf("Deleting managed service %s from Kubernetes", service.Name)
	
	// Use the existing Kubernetes deletion utility (it should work for managed services too)
	err := utils.DeleteKubernetesResources(service)
	if err != nil {
		return fmt.Errorf("failed to delete Kubernetes resources: %v", err)
	}
	
	log.Printf("Managed service %s deleted from Kubernetes", service.Name)
	return nil
}
// checkIfRedeploymentNeeded determines if changes require redeployment
func (s *ManagedServiceService) checkIfRedeploymentNeeded(existing, updated models.Service) bool {
	// Changes that require redeployment
	return existing.Version != updated.Version ||
		existing.CPULimit != updated.CPULimit ||
		existing.MemoryLimit != updated.MemoryLimit ||
		existing.StorageSize != updated.StorageSize ||
		existing.EnvironmentID != updated.EnvironmentID ||
		existing.CustomDomain != updated.CustomDomain
}