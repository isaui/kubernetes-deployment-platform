package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
	"github.com/pendeploy-simple/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentService struct {
	serviceRepo    *repositories.ServiceRepository
	deploymentRepo *repositories.DeploymentRepository
	registryRepo   *repositories.RegistryRepository
}

func NewDeploymentService() *DeploymentService {
	return &DeploymentService{
		serviceRepo:    repositories.NewServiceRepository(),
		deploymentRepo: repositories.NewDeploymentRepository(),
		registryRepo:   repositories.NewRegistryRepository(),
	}
}

func (s *DeploymentService) CreateGitDeployment(request dto.GitDeployRequest) (dto.GitDeployResponse, error) {
	// Get service details to access repo URL and branch
	service, err := s.serviceRepo.FindByID(request.ServiceID)
	if err != nil {
		log.Println("Error fetching service details:", err)
		return dto.GitDeployResponse{}, err
	}
	// Validate the service ID and API key
	isValid, err := utils.ValidateServiceDeployment(service, request.APIKey)
	if err != nil {
		log.Println("Error validating service deployment:", err)
		return dto.GitDeployResponse{}, err
	}
	if !isValid {
		return dto.GitDeployResponse{}, fmt.Errorf("unauthorized: invalid API key")
	}

	deployment, err := s.deploymentRepo.Create(models.Deployment{
		ServiceID:     service.ID,
		Status:        "building",
		CommitSHA:     request.CommitID,
		CommitMessage: request.CommitMessage,
	})
	if err != nil {
		log.Println("Error creating deployment:", err)
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		return dto.GitDeployResponse{}, err
	}

	registry, err := s.registryRepo.FindDefault()
	if err != nil {
		log.Println("Error fetching registry details:", err)
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		return dto.GitDeployResponse{}, err
	}

	go s.ProcessGitDeployment(deployment, service, registry, request.CallbackUrl)

	return dto.GitDeployResponse{
		DeploymentID: deployment.ID,
		ServiceID:    service.ID,
		Status:       "building",
		JobName:      utils.GetJobName(service.ID, deployment.ID),
		Message:      "Deployment started",
		CreatedAt:    deployment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *DeploymentService) ProcessGitDeployment(deployment models.Deployment,
	 service models.Service, registry models.Registry, callbackUrl string) error {

	log.Println("Processing Git deployment for service:", service.Name)
	// Start building the image
	image, err := utils.BuildFromGit(deployment, service, registry)
	if err != nil {
		log.Println("Error building image:", err)
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}

	
	err = s.deploymentRepo.UpdateImage(deployment.ID, image)
	if err != nil {
		log.Println("Error updating image:", err)
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}

	updatedService, err := s.DeployToKubernetes(image, service)
	if err != nil {
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		s.serviceRepo.Update(*updatedService)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}
	
	log.Println("Deployment successful for service:", service.Name)
	s.serviceRepo.Update(*updatedService)
	s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusSuccess)
	// Call webhook with success status if URL exists
	if callbackUrl != "" {
		go utils.SendWebhookNotification(callbackUrl, deployment.ID, "running", "")
	}
	return nil
}

func (s *DeploymentService) DeployToKubernetes(imageUrl string, service models.Service) (*models.Service, error) {
	// Deploy all Kubernetes resources (Deployment, Service, Ingress) atomically
	log.Println("Deploying to Kubernetes for service:", service.Name)
	updatedService, err := utils.DeployToKubernetesAtomically(imageUrl, service)
	if err != nil {
		log.Println("Error deploying to Kubernetes:", err)
		return nil, fmt.Errorf("failed to deploy to Kubernetes: %v", err)
	}
	
	return updatedService, nil
}

func (s *DeploymentService) GetDeploymentByID(id string) (*dto.DeploymentResponse, error) {
	deployment, err := s.deploymentRepo.FindByID(id)
	if err != nil {
		log.Println("Error fetching deployment details:", err)
		return nil, err
	}
	
	response := dto.NewDeploymentResponseFromModel(deployment)
	return &response, nil
}

func (s *DeploymentService) GetResourceStatus(serviceID string) (*dto.ResourceStatusResponse, error) {
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		log.Println("Error fetching service details:", err)
		return nil, err
	}
	
	resourceMap, err := utils.GetKubernetesResourceStatus(service)
	if err != nil {
		log.Println("Error fetching Kubernetes resource status:", err)
		return nil, err
	}
	
	// Konversi map menjadi struktur dto.ResourceStatusResponse
	response := &dto.ResourceStatusResponse{}
	
	// Ekstrak data deployment
	if deploymentData, ok := resourceMap["deployment"].(map[string]interface{}); ok {
		response.Deployment = &dto.DeploymentStatusInfo{
			Name: utils.GetString(deploymentData, "name"),
			ReadyReplicas: utils.GetInt32(deploymentData, "readyReplicas"),
			AvailableReplicas: utils.GetInt32(deploymentData, "availableReplicas"),
			Replicas: utils.GetInt32(deploymentData, "replicas"),
			Image: utils.GetString(deploymentData, "image"),
		}
	}
	
	// Ekstrak data service
	if serviceData, ok := resourceMap["service"].(map[string]interface{}); ok {
		response.Service = &dto.ServiceStatusInfo{
			Name: utils.GetString(serviceData, "name"),
			Type: utils.GetString(serviceData, "type"),
			ClusterIP: utils.GetString(serviceData, "clusterIP"),
			Ports: utils.GetString(serviceData, "ports"),
		}
	}
	
	// Ekstrak data ingress
	if ingressData, ok := resourceMap["ingress"].(map[string]interface{}); ok {
		hosts := []string{}
		if rulesData, ok := ingressData["hosts"].([]interface{}); ok {
			for _, rule := range rulesData {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					if host, ok := ruleMap["host"].(string); ok {
						hosts = append(hosts, host)
					}
				}
			}
		}
		
		response.Ingress = &dto.IngressStatusInfo{
			Name: utils.GetString(ingressData, "name"),
			Hosts: hosts,
			TLS: utils.GetBool(ingressData, "tls"),
			Status: utils.GetString(ingressData, "status"),
		}
	}
	
	// Ekstrak data HPA
	if hpaData, ok := resourceMap["hpa"].(map[string]interface{}); ok {
		response.HPA = &dto.HPAStatusInfo{
			Name: utils.GetString(hpaData, "name"),
			MinReplicas: utils.GetInt32(hpaData, "minReplicas"),
			MaxReplicas: utils.GetInt32(hpaData, "maxReplicas"),
			CurrentReplicas: utils.GetInt32(hpaData, "currentReplicas"),
			TargetCPU: utils.GetInt32(hpaData, "targetCPU"),
			CurrentCPU: utils.GetInt32(hpaData, "currentCPU"),
		}
	}
	
	return response, nil
}


func (s *DeploymentService) GetServiceBuildLogsRealtime(deploymentID string, w http.ResponseWriter) error {
	// Get deployment details from database
	deployment, err := s.deploymentRepo.FindByID(deploymentID)
	if err != nil {
		return fmt.Errorf("deployment not found: %v", err)
	}
	
	// Get service details
	service, err := s.serviceRepo.FindByID(deployment.ServiceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}
	
	// Get job name
	jobName := utils.GetJobName(service.ID, deployment.ID)
	namespace := utils.GetJobNamespace()
	
	// Flush to ensure headers are sent
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	flusher.Flush()
	
	// Write initial connection message
	utils.WriteSSEMessage(w, fmt.Sprintf("Connected to build logs for job %s", jobName))
	flusher.Flush()
	
	// Get job pods
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		utils.WriteSSEMessage(w, fmt.Sprintf("Error finding job pods: %v", err))
		flusher.Flush()
		return err
	}
	
	if len(podList.Items) == 0 {
		utils.WriteSSEMessage(w, "No pods found for this job yet")
		flusher.Flush()
		
		// Wait for a pod to appear for up to 60 seconds
		for i := 0; i < 12; i++ {
			time.Sleep(5 * time.Second)
			
			podList, err = k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", jobName),
			})
			
			if err == nil && len(podList.Items) > 0 {
				utils.WriteSSEMessage(w, "Pod found, connecting to logs")
				flusher.Flush()
				break
			}
			
			if i == 11 {
				utils.WriteSSEMessage(w, "Timeout waiting for job pod to appear")
				flusher.Flush()
				return fmt.Errorf("timeout waiting for job pod")
			}
		}
	}
	
	// Get the pod
	podName := podList.Items[0].Name
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	// Get pod details to discover containers dynamically
	pod, err := k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		utils.WriteSSEMessage(w, fmt.Sprintf("Error getting pod details: %v", err))
		flusher.Flush()
		return err
	}
	
	// Build container list dynamically: init containers first, then main containers
	var containersToStream []string
	
	// Add init containers (e.g., git-clone)
	for _, initContainer := range pod.Spec.InitContainers {
		containersToStream = append(containersToStream, initContainer.Name)
	}
	
	// Add main containers (e.g., kaniko-executor, buildah-executor, cnb-pack-executor)
	for _, container := range pod.Spec.Containers {
		containersToStream = append(containersToStream, container.Name)
	}
	
	utils.WriteSSEMessage(w, fmt.Sprintf("Found containers to stream: %v", containersToStream))
	flusher.Flush()
	
	// Stream logs from each container in sequence
	for i, containerName := range containersToStream {
		utils.WriteSSEMessage(w, fmt.Sprintf("=== Streaming logs from %s container ===", containerName))
		flusher.Flush()
		
		// Wait for container to be ready before streaming logs
		if err := s.waitForContainerReady(ctx, k8sClient, namespace, podName, containerName, w, flusher); err != nil {
			utils.WriteSSEMessage(w, fmt.Sprintf("Container %s not ready: %v", containerName, err))
			continue
		}
		
		// Stream logs from this container
		if err := s.streamContainerLogs(ctx, k8sClient, namespace, podName, containerName, w, flusher); err != nil {
			utils.WriteSSEMessage(w, fmt.Sprintf("Error streaming logs from %s: %v", containerName, err))
		}
		
		// Add separator between containers (except for the last one)
		if i < len(containersToStream)-1 {
			utils.WriteSSEMessage(w, fmt.Sprintf("=== End of %s logs ===", containerName))
			flusher.Flush()
		}
	}
	
	// Final completion message
	utils.WriteSSEMessage(w, "=== Build log streaming completed ===")
	flusher.Flush()
	
	return nil
}

// waitForContainerReady waits for a container to be ready before streaming logs
func (s *DeploymentService) waitForContainerReady(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName, containerName string, w http.ResponseWriter, flusher http.Flusher) error {
	maxWait := 2 * time.Minute
	checkInterval := 2 * time.Second
	timeout := time.After(maxWait)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for container %s to be ready", containerName)
		case <-ticker.C:
			pod, err := k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				continue
			}
			
			// Check if container is ready (for init containers, check if completed)
			for _, initContainerStatus := range pod.Status.InitContainerStatuses {
				if initContainerStatus.Name == containerName {
					if initContainerStatus.State.Running != nil || initContainerStatus.State.Terminated != nil {
						return nil // Container is running or has completed
					}
				}
			}
			
			// Check main containers
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == containerName {
					if containerStatus.State.Running != nil {
						return nil // Container is running
					}
					if containerStatus.State.Terminated != nil {
						return nil // Container has completed
					}
				}
			}
		}
	}
}

// streamContainerLogs streams logs from a specific container
func (s *DeploymentService) streamContainerLogs(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName, containerName string, w http.ResponseWriter, flusher http.Flusher) error {
	// Try to stream logs from this container
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		Follow:    true,
		TailLines: int64Ptr(1000), // Last 1000 lines to avoid overwhelming
	})
	
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("error opening log stream: %v", err)
	}
	defer stream.Close()
	
	// Read logs and stream them as SSE events
	reader := bufio.NewReader(stream)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					utils.WriteSSEMessage(w, fmt.Sprintf("Log stream ended for %s", containerName))
					return nil
				}
				return fmt.Errorf("error reading logs: %v", err)
			}
			
			// Write each line as an SSE event
			utils.WriteSSEData(w, string(line))
			flusher.Flush()
		}
	}
}


func (s *DeploymentService) GetServiceRuntimeLogsRealtime(serviceID string, w http.ResponseWriter) error {
	// Get service details
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Get the immutable resource name for the deployment
	deploymentResourceName := utils.GetResourceName(service)
	
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}
	
	// Namespace matches the service environment
	namespace := service.EnvironmentID
	
	// Flush to ensure headers are sent
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	flusher.Flush()
	
	// Write initial connection message
	utils.WriteSSEMessage(w, fmt.Sprintf("Connected to runtime logs for deployment %s", deploymentResourceName))
	flusher.Flush()
	
	// Get pods for the deployment
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentResourceName),
	})
	if err != nil {
		utils.WriteSSEMessage(w, fmt.Sprintf("Error finding deployment pods: %v", err))
		flusher.Flush()
		return err
	}
	
	if len(podList.Items) == 0 {
		utils.WriteSSEMessage(w, "No pods found for this deployment")
		flusher.Flush()
		return fmt.Errorf("no pods found")
	}
	
	// Get the first pod (we can enhance this later to stream from all pods or allow pod selection)
	podName := podList.Items[0].Name
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	// Stream logs from the main container (assuming container name is app)
	containerName := "app" 
	
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		Follow:    true,
		TailLines: int64Ptr(100), // Start with the last 100 lines
	})
	
	stream, err := req.Stream(ctx)
	if err != nil {
		utils.WriteSSEMessage(w, fmt.Sprintf("Error opening log stream: %v", err))
		flusher.Flush()
		return err
	}
	defer stream.Close()
	
	// Read logs and stream them as SSE events
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				utils.WriteSSEMessage(w, "Log stream ended")
				break
			}
			utils.WriteSSEMessage(w, fmt.Sprintf("Error reading logs: %v", err))
			break
		}
		
		// Write each line as an SSE event
		utils.WriteSSEData(w, string(line))
		flusher.Flush()
	}
	
	return nil
}

