package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
	"github.com/pendeploy-simple/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	registryNamespace = "registry"
	registryTimeout   = 10 * time.Minute
)

// RegistryService handles business logic for registries
type RegistryService struct {
	registryRepo *repositories.RegistryRepository
	kubeClient   *kubernetes.Client
	depService   *RegistryDependencyService
}

// NewRegistryService creates a new registry service instance
func NewRegistryService() *RegistryService {
	client, err := kubernetes.NewClient()
	if err != nil {
		// Log error but continue - operations requiring K8s will fail gracefully
		fmt.Printf("Warning: Could not create Kubernetes client: %v\n", err)
	}
	
	return &RegistryService{
		registryRepo: repositories.NewRegistryRepository(),
		kubeClient:   client,
		depService:   NewRegistryDependencyService(),
	}
}

// ListRegistries retrieves registries with pagination, filtering and sorting
func (s *RegistryService) ListRegistries(filter dto.RegistryFilter) (dto.RegistryListResponse, error) {
	var response dto.RegistryListResponse
	
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
	
	// Get registries from repository
	registries, total, err := s.registryRepo.FindWithPagination(
		filter.Page, 
		filter.PageSize, 
		filter.SortBy, 
		filter.SortOrder,
		filter.Search,
		filter.OnlyActive,
	)
	
	if err != nil {
		return response, err
	}
	
	// Convert to DTO
	registryResponses := make([]dto.RegistryResponse, len(registries))
	for i, registry := range registries {
		registryResponses[i] = convertRegistryToResponse(registry)
	}
	
	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))
	
	// Build response
	response = dto.RegistryListResponse{
		Registries: registryResponses,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}
	
	return response, nil
}

// GetRegistryByID retrieves a registry by ID
func (s *RegistryService) GetRegistryByID(id string) (dto.RegistryResponse, error) {
	registry, err := s.registryRepo.FindByID(id)
	if err != nil {
		return dto.RegistryResponse{}, err
	}
	
	return convertRegistryToResponse(registry), nil
}

// CreateRegistry creates a new registry and initiates deployment in Kubernetes
func (s *RegistryService) CreateRegistry(req dto.CreateRegistryRequest) (dto.RegistryResponse, error) {
	// Create registry model
	registry := models.Registry{
		Name:      req.Name,
		IsDefault: req.IsDefault,
		IsActive:  true,
		Status:    models.RegistryStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save to database
	createdRegistry, err := s.registryRepo.Create(registry)
	if err != nil {
		return dto.RegistryResponse{}, err
	}
	
	// Start async deployment
	go s.deployRegistryInKubernetes(createdRegistry.ID)
	
	return convertRegistryToResponse(createdRegistry), nil
}

// UpdateRegistry updates an existing registry
func (s *RegistryService) UpdateRegistry(id string, req dto.UpdateRegistryRequest) (dto.RegistryResponse, error) {
	// Get existing registry
	registry, err := s.registryRepo.FindByID(id)
	if err != nil {
		return dto.RegistryResponse{}, err
	}
	// Update fields if provided
	if req.Name != "" {
		registry.Name = req.Name
	}
	registry.IsDefault = req.IsDefault
	registry.UpdatedAt = time.Now()
	
	// Save to database
	if err := s.registryRepo.Update(registry); err != nil {
		return dto.RegistryResponse{}, err
	}
	
	// Update Kubernetes resources if needed
	go s.updateRegistryInKubernetes(id)
	
	return convertRegistryToResponse(registry), nil
}

// DeleteRegistry removes a registry
func (s *RegistryService) DeleteRegistry(id string) error {
	// Check if registry exists
	registry, err := s.registryRepo.FindByID(id)
	if err != nil {
		return err
	}
	
	log.Printf("Deleting registry with ID %s and BuildPodName %s", id, registry.BuildPodName)
	
	// Delete from Kubernetes first
	if err := s.deleteRegistryFromKubernetes(registry.ID); err != nil {
		// Return error and do NOT delete from database to preserve tracking ability
		log.Printf("Error: Failed to delete registry from Kubernetes: %v\n", err)
		return fmt.Errorf("failed to delete Kubernetes resources: %v", err)
	}
	
	// Only delete from database if Kubernetes deletion succeeded
	return s.registryRepo.Delete(id)
}

// GetRegistryDetails retrieves detailed registry information including Kubernetes data
func (s *RegistryService) GetRegistryDetails(id string) (dto.RegistryDetailsResponse, error) {
	// Get basic registry info
	registry, err := s.registryRepo.FindByID(id)
	if err != nil {
		return dto.RegistryDetailsResponse{}, err
	}
	
	// Create response with basic info
	response := dto.RegistryDetailsResponse{
		Registry:    convertRegistryToResponse(registry),
		Credentials: &dto.RegistryCredentials{URL: registry.URL},
		IsHealthy:   true, // Default to true
		Images:      []dto.RegistryImageInfo{}, // Initialize empty slice
		ImagesCount: 0,
		LastSynced:  nil,
	}
	
	
	// Only fetch Kubernetes data if client is available
	if s.kubeClient != nil {
		// Get pod status
		pod, err := s.getRegistryPod(registry.BuildPodName)
		if err == nil && pod != nil {
			response.KubeStatus = string(pod.Status.Phase)
		}
		
		// Create credentials response using registry data from database
		if registry.URL != "" {

			log.Printf("Registry credentials - URL: %s, Username: %s", registry.URL, "admin")
			
			// Use the UPDATED RegistryAPI - with HTTPS support
			apiClient, err := utils.NewRegistryAPIFromRegistry(registry.URL)
			if err != nil {
				log.Printf("Failed to create registry API client: %v", err)
				response.IsHealthy = false
			} else {
				// Test connection first
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				
				err = utils.TestConnection(ctx, apiClient)
				if err != nil {
					log.Printf("Registry connection test failed: %v", err)
					response.IsHealthy = false
				} else {
					log.Printf("Registry connection test successful!")
					
					// Get list of images with details
					images, err := utils.ListImages(ctx, apiClient)
					if err != nil {
						log.Printf("Failed to list images: %v", err)
						response.IsHealthy = false
					} else {
						// Update response with image information
						response.Images = images
						response.ImagesCount = len(images)
						
						// Calculate total size
						var totalSize int64
						for _, img := range images {
							totalSize += img.Size
						}
						response.Size = totalSize
						
						// Set last synced time to now
						now := time.Now()
						response.LastSynced = &now
						
						log.Printf("Successfully retrieved %d images, total size: %d bytes", len(images), totalSize)
					}
				}
			}
		} else {
			log.Printf("Registry URL is empty")
			response.IsHealthy = false
		}
	}
	
	return response, nil
}

// SetupRegistryDependencies manually sets up dependencies for an existing registry
func (s *RegistryService) SetupRegistryDependencies(registryID string) error {
	if s.kubeClient == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}
	
	// Get registry
	registry, err := s.registryRepo.FindByID(registryID)
	if err != nil {
		return fmt.Errorf("failed to get registry: %v", err)
	}
	
	// Setup dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	log.Printf("Manually setting up dependencies for registry %s", registryID)
	return s.depService.SetupRegistryDependencies(ctx, registry)
}

// ValidateDependencies checks if all required images are available in the registry
func (s *RegistryService) ValidateDependencies(registryID string) ([]string, error) {
	// Get registry
	registry, err := s.registryRepo.FindByID(registryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry: %v", err)
	}
	
	// Validate dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	return s.depService.ValidateDependencies(ctx, registry)
}

// StreamRegistryBuildLogs streams build logs from a registry deployment pod
func (s *RegistryService) StreamRegistryBuildLogs(ctx context.Context, registryID string, w io.Writer) error {
	log.Println("Starting StreamRegistryBuildLogs for registry ID:", registryID)

	if s.kubeClient == nil {
		log.Println("Error: Kubernetes client is not initialized")
		return fmt.Errorf("kubernetes client is not initialized")
	}
	
	// Get registry
	registry, err := s.registryRepo.FindByID(registryID)
	if err != nil {
		log.Println("Error finding registry by ID:", err)
		return err
	}
	
	log.Printf("Registry found: ID=%s, BuildPodName=%s", registryID, registry.BuildPodName)
	
	// If no pod name is stored yet, we need to wait for it
	if registry.BuildPodName == "" {
		// Write initial message
		fmt.Fprintf(w, "Waiting for registry pod to be created...\n")
		
		// Poll until pod is created or timeout
		err = wait.PollImmediate(time.Second, time.Minute*2, func() (bool, error) {
			// Reload registry data
			freshRegistry, err := s.registryRepo.FindByID(registryID)
			if err != nil {
				return false, err
			}
			
			// Check if pod name is available now
			if freshRegistry.BuildPodName != "" {
				registry = freshRegistry
				return true, nil
			}
			
			// Not ready yet
			fmt.Fprintf(w, ".")
			return false, nil
		})
		
		if err != nil {
			return fmt.Errorf("timeout waiting for registry pod: %v", err)
		}
		
		fmt.Fprintf(w, "\nPod created: %s\n", registry.BuildPodName)
	}
	
	// First fetch historical logs (without following)
	fmt.Fprintf(w, "Fetching log history...\n")
	
	// Get the namespace from utils
	registryNamespace := utils.RegistryNamespace
	log.Printf("Using namespace: %s for pod: %s", registryNamespace, registry.BuildPodName)
	
	// Set up historical log options
	historicalLogOpts := &corev1.PodLogOptions{
		Follow:    false, // Don't follow for historical logs
		Container: "registry", // Container name verified via K8s API
	}
	
	// Log debug info
	log.Printf("Debug - Attempting to get logs for: Namespace=%s, Pod=%s, Container=%s", 
		registryNamespace, registry.BuildPodName, "registry")
	
	// Get historical logs
	histReq := s.kubeClient.Clientset.CoreV1().Pods(registryNamespace).GetLogs(registry.BuildPodName, historicalLogOpts)
	histLogs, err := histReq.Stream(ctx)
	if err != nil {
		log.Printf("Error streaming historical logs: %v", err)
		fmt.Fprintf(w, "data: Warning: Could not fetch historical logs: %v\n\n", err)
	} else {
		defer histLogs.Close()
		
		// Copy historical logs to writer with a clear header
		fmt.Fprintf(w, "data: \n--- Begin Log History ---\n\n")
		
		// Read and write logs in SSE format
		buffer := make([]byte, 4096)
		for {
			n, err := histLogs.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading historical logs: %v", err)
				}
				break
			}
			
			if n > 0 {
				// Format each line as an SSE event
				fmt.Fprintf(w, "data: %s\n\n", buffer[:n])
			}
		}
		
		fmt.Fprintf(w, "data: \n--- End Log History ---\n\n")
	}
	
	// Now set up streaming for new logs
	fmt.Fprintf(w, "data: Now streaming live logs...\n\n")
	
	// Set up live log streaming options
	livePodLogOpts := &corev1.PodLogOptions{
		Follow:    true,
		Container: "registry",  // Adjust container name based on your setup
		// Don't use TailLines here as we want to start from where the historical logs ended
	}
	
	// Get live logs stream
	log.Printf("Attempting to stream live logs for pod %s in namespace %s", registry.BuildPodName, registryNamespace)
	liveReq := s.kubeClient.Clientset.CoreV1().Pods(registryNamespace).GetLogs(registry.BuildPodName, livePodLogOpts)
	liveLogs, err := liveReq.Stream(ctx)
	if err != nil {
		log.Printf("Error opening live log stream: %v", err)
		return fmt.Errorf("error opening live log stream: %v", err)
	}
	defer liveLogs.Close()
	
	// Stream live logs to writer in SSE format
	buffer := make([]byte, 4096)
	for {
		n, err := liveLogs.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading live logs: %v", err)
				return err
			}
			break
		}
		
		if n > 0 {
			// Format as SSE event
			fmt.Fprintf(w, "data: %s\n\n", buffer[:n])
			// Flush to ensure data is sent immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
	return err
}

// deployRegistryInKubernetes deploys a registry in Kubernetes asynchronously
func (s *RegistryService) deployRegistryInKubernetes(registryID string) {
	if s.kubeClient == nil {
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, "Kubernetes client is not initialized")
		return
	}
	
	// Update registry status to building
	s.updateRegistryStatus(registryID, models.RegistryStatusBuilding, "")
	
	// Get registry data
	registry, err := s.registryRepo.FindByID(registryID)
	if err != nil {
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, fmt.Sprintf("Failed to get registry: %v", err))
		return
	}
	
	// Create context with timeout for deployment operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // Extended timeout for dependencies
	defer cancel()
	
	// Create a registry deployer
	deployer := NewRegistryDeployer(s.kubeClient.Clientset)
	
	// Deploy the registry using our utility
	podName, registryURL, err := deployer.DeployRegistry(ctx, registry)
	if err != nil {
		// Log the error
		fmt.Printf("Error deploying registry %s: %v\n", registry.ID, err)
		
		// Update registry status to failed
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, fmt.Sprintf("Failed to deploy registry: %v", err))
		return
	}
	
	// Update registry with pod name and URL
	registry.BuildPodName = podName
	registry.URL = registryURL
	registry.IsActive = true
	
	fmt.Printf("Registry URL set to: %s\n", registryURL)
	
	// Update registry in database
	s.registryRepo.Update(registry)
	
	// Wait a moment to ensure the pod is fully running
	time.Sleep(10 * time.Second)

	s.updateRegistryStatus(registryID, models.RegistryStatusReady, "Registry ready")
}

// updateRegistryInKubernetes updates registry configuration in Kubernetes
func (s *RegistryService) updateRegistryInKubernetes(registryID string) {
	if s.kubeClient == nil {
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, "Kubernetes client is not initialized")
		return
	}
	
	// Update registry status to building
	s.updateRegistryStatus(registryID, models.RegistryStatusBuilding, "")
	
	// Get registry data
	registry, err := s.registryRepo.FindByID(registryID)
	if err != nil {
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, fmt.Sprintf("Failed to get registry: %v", err))
		return
	}
	
	// Create context with timeout for update operations
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	
	// Create a registry deployer
	deployer := NewRegistryDeployer(s.kubeClient.Clientset)
	
	// Update the registry using our utility
	err = deployer.UpdateRegistry(ctx, registry)
	if err != nil {
		// Log the error
		fmt.Printf("Error updating registry %s: %v\n", registry.ID, err)
		
		// Update registry status to failed
		s.updateRegistryStatus(registryID, models.RegistryStatusFailed, fmt.Sprintf("Failed to update registry: %v", err))
		return
	}
	
	// Update status to ready
	s.updateRegistryStatus(registryID, models.RegistryStatusReady, "")
}

// deleteRegistryFromKubernetes deletes registry resources from Kubernetes
func (s *RegistryService) deleteRegistryFromKubernetes(registryID string) error {
	if s.kubeClient == nil || registryID == "" {
		return nil // Nothing to delete
	}
	
	// Create context with timeout for delete operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Extended timeout for cleanup
	defer cancel()
	
	// Cleanup dependency jobs first
	log.Printf("Cleaning up dependency jobs for registry %s", registryID)
	err := s.depService.CleanupDependencyJobs(ctx, registryID, 0) // Delete all jobs
	if err != nil {
		log.Printf("Warning: Failed to cleanup dependency jobs for registry %s: %v", registryID, err)
		// Continue anyway - dependency cleanup failure shouldn't block registry deletion
	} else {
		log.Printf("Successfully cleaned up dependency jobs for registry %s", registryID)
	}
	
	// Create a registry deployer
	deployer := NewRegistryDeployer(s.kubeClient.Clientset)
	
	// Use the deployer to delete all registry resources
	err = deployer.DeleteRegistry(ctx, registryID)
	if err != nil {
		log.Printf("Failed to delete registry resources for %s: %v", registryID, err)
		return err
	}
	
	log.Printf("Successfully deleted registry resources for %s", registryID)
	return nil
}

// getRegistryPod gets the registry pod from Kubernetes
func (s *RegistryService) getRegistryPod(podName string) (*corev1.Pod, error) {
	if s.kubeClient == nil || podName == "" {
		return nil, fmt.Errorf("kubernetes client not initialized or pod name empty")
	}
	
	return s.kubeClient.Clientset.CoreV1().Pods(registryNamespace).Get(
		context.Background(),
		podName,
		metav1.GetOptions{},
	)
}

// updateRegistryStatus updates the status of a registry in the database
func (s *RegistryService) updateRegistryStatus(id string, status models.RegistryStatus, message string) {
	registry, err := s.registryRepo.FindByID(id)
	if err != nil {
		fmt.Printf("Error updating registry status: %v\n", err)
		return
	}
	
	registry.Status = status
	registry.UpdatedAt = time.Now()
	
	if err := s.registryRepo.Update(registry); err != nil {
		fmt.Printf("Error saving registry status: %v\n", err)
	}
	
	// Log status update with message
	if message != "" {
		log.Printf("Registry %s status updated to %s: %s", id, status, message)
	} else {
		log.Printf("Registry %s status updated to %s", id, status)
	}
}

// convertRegistryToResponse converts a registry model to a DTO response
func convertRegistryToResponse(registry models.Registry) dto.RegistryResponse {
	return dto.RegistryResponse{
		ID:        registry.ID,
		Name:      registry.Name,
		URL:       registry.URL,
		IsDefault: registry.IsDefault,
		IsActive:  registry.IsActive,
		Status:    registry.Status,
		CreatedAt: registry.CreatedAt,
		UpdatedAt: registry.UpdatedAt,
	}
}