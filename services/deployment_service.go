package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		return dto.GitDeployResponse{}, err
	}
	// Validate the service ID and API key
	isValid, err := utils.ValidateServiceDeployment(service, request.APIKey)
	if err != nil {
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
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		return dto.GitDeployResponse{}, err
	}

	registry, err := s.registryRepo.FindDefault()
	if err != nil {
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

	// Start building the image
	image, err := utils.BuildFromGit(deployment, service, registry)
	if err != nil {
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}
	
	err = s.deploymentRepo.UpdateImage(deployment.ID, image)
	if err != nil {
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}

	err = s.DeployToKubernetes(image, service)
	if err != nil {
		s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusFailed)
		// Call webhook with failure status if URL exists
		if callbackUrl != "" {
			go utils.SendWebhookNotification(callbackUrl, deployment.ID, "failed", err.Error())
		}
		return err
	}
	
	s.deploymentRepo.UpdateStatus(deployment.ID, models.DeploymentStatusSuccess)
	// Call webhook with success status if URL exists
	if callbackUrl != "" {
		go utils.SendWebhookNotification(callbackUrl, deployment.ID, "running", "")
	}
	return nil
}

func (s *DeploymentService) DeployToKubernetes(imageUrl string, service models.Service) error {
	// Deploy all Kubernetes resources (Deployment, Service, Ingress) atomically
	err := utils.DeployToKubernetesAtomically(imageUrl, service)
	if err != nil {
		return fmt.Errorf("failed to deploy to Kubernetes: %v", err)
	}
	
	return nil
}

func (s *DeploymentService) GetDeploymentByID(id string) (*dto.DeploymentResponse, error) {
	deployment, err := s.deploymentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	response := dto.NewDeploymentResponseFromModel(deployment)
	return &response, nil
}

func (s *DeploymentService) GetResourceStatus(serviceID string) (*dto.ResourceStatusResponse, error) {
	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return nil, err
	}
	
	resourceMap, err := utils.GetKubernetesResourceStatus(service)
	if err != nil {
		return nil, err
	}
	
	// Konversi map menjadi struktur dto.ResourceStatusResponse
	response := &dto.ResourceStatusResponse{}
	
	// Ekstrak data deployment
	if deploymentData, ok := resourceMap["deployment"].(map[string]interface{}); ok {
		response.Deployment = &dto.DeploymentStatusInfo{
			Name: getString(deploymentData, "name"),
			ReadyReplicas: getInt32(deploymentData, "readyReplicas"),
			AvailableReplicas: getInt32(deploymentData, "availableReplicas"),
			Replicas: getInt32(deploymentData, "replicas"),
			Image: getString(deploymentData, "image"),
		}
	}
	
	// Ekstrak data service
	if serviceData, ok := resourceMap["service"].(map[string]interface{}); ok {
		response.Service = &dto.ServiceStatusInfo{
			Name: getString(serviceData, "name"),
			Type: getString(serviceData, "type"),
			ClusterIP: getString(serviceData, "clusterIP"),
			Ports: getString(serviceData, "ports"),
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
			Name: getString(ingressData, "name"),
			Hosts: hosts,
			TLS: getBool(ingressData, "tls"),
			Status: getString(ingressData, "status"),
		}
	}
	
	// Ekstrak data HPA
	if hpaData, ok := resourceMap["hpa"].(map[string]interface{}); ok {
		response.HPA = &dto.HPAStatusInfo{
			Name: getString(hpaData, "name"),
			MinReplicas: getInt32(hpaData, "minReplicas"),
			MaxReplicas: getInt32(hpaData, "maxReplicas"),
			CurrentReplicas: getInt32(hpaData, "currentReplicas"),
			TargetCPU: getInt32(hpaData, "targetCPU"),
			CurrentCPU: getInt32(hpaData, "currentCPU"),
		}
	}
	
	return response, nil
}

// Helper functions untuk ekstraksi data dari map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt32(data map[string]interface{}, key string) int32 {
	switch val := data[key].(type) {
	case int32:
		return val
	case int:
		return int32(val)
	case int64:
		return int32(val)
	case float32:
		return int32(val)
	case float64:
		return int32(val)
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
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
	writeSSEMessage(w, fmt.Sprintf("Connected to build logs for job %s", jobName))
	flusher.Flush()
	
	// Get job pods
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		writeSSEMessage(w, fmt.Sprintf("Error finding job pods: %v", err))
		flusher.Flush()
		return err
	}
	
	if len(podList.Items) == 0 {
		writeSSEMessage(w, "No pods found for this job yet")
		flusher.Flush()
		
		// Wait for a pod to appear for up to 60 seconds
		for i := 0; i < 12; i++ {
			time.Sleep(5 * time.Second)
			
			podList, err = k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", jobName),
			})
			
			if err == nil && len(podList.Items) > 0 {
				writeSSEMessage(w, "Pod found, connecting to logs")
				flusher.Flush()
				break
			}
			
			if i == 11 {
				writeSSEMessage(w, "Timeout waiting for job pod to appear")
				flusher.Flush()
				return fmt.Errorf("timeout waiting for job pod")
			}
		}
	}
	
	// Get the pod
	podName := podList.Items[0].Name
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	// Stream logs from the main container
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "nixpacks-builder",
		Follow:    true,
	})
	
	stream, err := req.Stream(ctx)
	if err != nil {
		writeSSEMessage(w, fmt.Sprintf("Error opening log stream: %v", err))
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
				writeSSEMessage(w, "Log stream ended")
				break
			}
			writeSSEMessage(w, fmt.Sprintf("Error reading logs: %v", err))
			break
		}
		
		// Write each line as an SSE event
		writeSSEData(w, string(line))
		flusher.Flush()
	}
	
	return nil
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
	writeSSEMessage(w, fmt.Sprintf("Connected to runtime logs for deployment %s", deploymentResourceName))
	flusher.Flush()
	
	// Get pods for the deployment
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentResourceName),
	})
	if err != nil {
		writeSSEMessage(w, fmt.Sprintf("Error finding deployment pods: %v", err))
		flusher.Flush()
		return err
	}
	
	if len(podList.Items) == 0 {
		writeSSEMessage(w, "No pods found for this deployment")
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
		writeSSEMessage(w, fmt.Sprintf("Error opening log stream: %v", err))
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
				writeSSEMessage(w, "Log stream ended")
				break
			}
			writeSSEMessage(w, fmt.Sprintf("Error reading logs: %v", err))
			break
		}
		
		// Write each line as an SSE event
		writeSSEData(w, string(line))
		flusher.Flush()
	}
	
	return nil
}

// Helper functions for SSE formatting
func writeSSEData(w io.Writer, data string) {
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func writeSSEMessage(w io.Writer, message string) {
	data := map[string]string{"message": message}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(w, "data: {\"message\": \"Error creating message\"}\n\n")
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}



