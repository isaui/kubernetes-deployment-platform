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
	"k8s.io/apimachinery/pkg/util/wait"
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
	log.Println("Starting build log streaming for deployment ID:", deploymentID)

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
	
	// Setup flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	
	// Get job details
	jobName := utils.GetJobName(service.ID, deployment.ID)
	namespace := utils.GetJobNamespace()
	
	log.Printf("Streaming logs for job: %s in namespace: %s", jobName, namespace)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	// Wait for job pod
	podName, err := s.waitForJobPod(ctx, k8sClient, namespace, jobName, w, flusher)
	if err != nil {
		return err
	}
	
	// Stream logs with dynamic container discovery
	return s.streamBuildLogsFromAllContainers(ctx, k8sClient, namespace, podName, w, flusher)
}

// streamBuildLogsFromAllContainers streams logs from all containers dynamically
func (s *DeploymentService) streamBuildLogsFromAllContainers(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName string, w http.ResponseWriter, flusher http.Flusher) error {
	// Get pod details to discover containers dynamically
	pod, err := k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	
	// Build container list: init containers first, then main containers
	var containers []string
	for _, initContainer := range pod.Spec.InitContainers {
		containers = append(containers, initContainer.Name)
	}
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}
	
	// Stream logs from each container
	for _, containerName := range containers {
		// Stream logs for this container
		err := s.streamSingleContainerLogs(ctx, k8sClient, namespace, podName, containerName, w, flusher)
		if err != nil {
			log.Printf("Warning: Error streaming logs from %s: %v", containerName, err)
		}
	}
	
	return nil
}

// streamSingleContainerLogs streams logs from a single container
func (s *DeploymentService) streamSingleContainerLogs(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName, containerName string, w http.ResponseWriter, flusher http.Flusher) error {
	// Wait for container to be ready
	err := s.waitForContainerReady(ctx, k8sClient, namespace, podName, containerName)
	if err != nil {
		return fmt.Errorf("container %s not ready: %v", containerName, err)
	}
	
	// Stream historical + live logs
	logOpts := &corev1.PodLogOptions{
		Follow:     true,
		Container:  containerName,
		TailLines:  int64Ptr(100),
		Timestamps: false, // Hapus timestamp container, biar frontend yang handle
	}
	
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("error opening log stream for %s: %v", containerName, err)
	}
	defer func() {
		if logs != nil {
			logs.Close()
		}
	}()
	
	// Stream logs line by line
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			utils.WriteSSEData(w, scanner.Text())
			flusher.Flush()
		}
	}
	
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading logs from %s: %v", containerName, err)
	}
	
	return nil
}

func (s *DeploymentService) GetServiceRuntimeLogsRealtime(serviceID string, w http.ResponseWriter) error {
	log.Println("Starting runtime log streaming for service ID:", serviceID)

	// Get service details
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %v", err)
	}
	
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}
	
	// Setup flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	
	deploymentResourceName := utils.GetResourceName(service)
	namespace := service.EnvironmentID
	
	log.Printf("Streaming runtime logs for deployment: %s in namespace: %s", deploymentResourceName, namespace)
	
	// Create context for streaming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle client disconnect
	if cn, ok := w.(http.CloseNotifier); ok {
		go func() {
			<-cn.CloseNotify()
			cancel()
		}()
	}
	
	// Get current pod (single attempt, no reconnect)
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentResourceName),
	})
	
	if err != nil {
		return err
	}
	
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found")
	}
	
	// Use first available pod
	pod := podList.Items[0]
	
	// Stream runtime logs (single attempt)
	return s.streamRuntimeLogsFromPod(ctx, k8sClient, namespace, pod.Name, w, flusher)
}

// streamRuntimeLogsFromPod streams runtime logs from a single pod
func (s *DeploymentService) streamRuntimeLogsFromPod(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName string, w http.ResponseWriter, flusher http.Flusher) error {
	// Get pod details to find container dynamically
	pod, err := k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %v", err)
	}
	
	// Use first container or default to "app"
	containerName := "app"
	if len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}
	
	// Stream logs with recent history + follow (realtime)
	logOpts := &corev1.PodLogOptions{
		Follow:     true,
		Container:  containerName,
		Timestamps: false, // Hapus timestamp container, biar frontend yang handle
		TailLines:  int64Ptr(50), // Recent 50 lines
	}
	
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("error opening runtime log stream: %v", err)
	}
	defer func() {
		if logs != nil {
			logs.Close()
		}
	}()
	
	// Stream logs line by line until context cancelled
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			utils.WriteSSEData(w, scanner.Text())
			flusher.Flush()
		}
	}
	
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading runtime logs: %v", err)
	}
	
	return nil
}

func (s *DeploymentService) waitForJobPod(ctx context.Context, k8sClient *kubernetes.Client, namespace, jobName string, w http.ResponseWriter, flusher http.Flusher) (string, error) {
	maxWait := 3 * time.Minute
	checkInterval := 5 * time.Second
	
	err := wait.PollImmediate(checkInterval, maxWait, func() (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", jobName),
			})
			
			if err != nil {
				return false, nil // Continue polling
			}
			
			if len(podList.Items) > 0 {
				return true, nil // Found pod
			}
			
			return false, nil // Continue polling
		}
	})
	
	if err != nil {
		return "", fmt.Errorf("timeout waiting for job pod: %v", err)
	}
	
	// Get pod name
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil || len(podList.Items) == 0 {
		return "", fmt.Errorf("failed to get job pod after wait")
	}
	
	return podList.Items[0].Name, nil
}

// waitForContainerReady waits for container to be ready
func (s *DeploymentService) waitForContainerReady(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName, containerName string) error {
	maxWait := 2 * time.Minute
	checkInterval := 2 * time.Second
	
	err := wait.PollImmediate(checkInterval, maxWait, func() (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			pod, err := k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return false, nil // Continue polling
			}
			
			// Check init containers
			for _, status := range pod.Status.InitContainerStatuses {
				if status.Name == containerName {
					if status.State.Running != nil || status.State.Terminated != nil {
						return true, nil
					}
				}
			}
			
			// Check main containers
			for _, status := range pod.Status.ContainerStatuses {
				if status.Name == containerName {
					if status.State.Running != nil || status.State.Terminated != nil {
						return true, nil
					}
				}
			}
			
			return false, nil // Continue polling
		}
	})
	
	return err
}

