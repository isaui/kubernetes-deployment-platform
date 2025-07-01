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
	"k8s.io/apimachinery/pkg/watch"
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
	
	// Handle client disconnect
	if cn, ok := w.(http.CloseNotifier); ok {
		go func() {
			<-cn.CloseNotify()
			cancel()
		}()
	}
	
	// Wait for job pod using watch
	podName, err := s.watchForJobPod(ctx, k8sClient, namespace, jobName, w, flusher)
	if err != nil {
		return err
	}
	
	// Stream logs from pod level
	return s.streamPodLogs(ctx, k8sClient, namespace, podName, w, flusher)
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
	
	// Watch for deployment pods and stream logs
	return s.watchAndStreamRuntimeLogs(ctx, k8sClient, namespace, deploymentResourceName, w, flusher)
}

// watchForJobPod uses watch API to wait for job pod
func (s *DeploymentService) watchForJobPod(ctx context.Context, k8sClient *kubernetes.Client, namespace, jobName string, w http.ResponseWriter, flusher http.Flusher) (string, error) {
	// Create watch for pods with job label
	watchOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
		Watch:         true,
	}
	
	watcher, err := k8sClient.Clientset.CoreV1().Pods(namespace).Watch(ctx, watchOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create pod watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Wait for pod with timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer timeoutCancel()
	
	for {
		select {
		case <-timeoutCtx.Done():
			return "", fmt.Errorf("timeout waiting for job pod")
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return "", fmt.Errorf("watch error: %v", event.Object)
			}
			
			if event.Type == watch.Added || event.Type == watch.Modified {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					continue
				}
				
				// Pod found, return name
				utils.WriteSSEData(w, fmt.Sprintf("Pod %s found, starting log stream...", pod.Name))
				flusher.Flush()
				return pod.Name, nil
			}
		}
	}
}

// watchAndStreamRuntimeLogs watches for deployment pods and streams their logs
func (s *DeploymentService) watchAndStreamRuntimeLogs(ctx context.Context, k8sClient *kubernetes.Client, namespace, deploymentName string, w http.ResponseWriter, flusher http.Flusher) error {
	// First, try to get current pod
	podList, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName),
	})
	
	var currentPod *corev1.Pod
	if err == nil && len(podList.Items) > 0 {
		// Get the newest running pod
		for i := range podList.Items {
			pod := &podList.Items[i]
			if pod.Status.Phase == corev1.PodRunning {
				currentPod = pod
				break
			}
		}
	}
	
	// If we have a current pod, stream its logs first
	if currentPod != nil {
		utils.WriteSSEData(w, fmt.Sprintf("Streaming logs from current pod: %s", currentPod.Name))
		flusher.Flush()
		
		// Stream current pod logs in background
		go func() {
			s.streamPodLogs(ctx, k8sClient, namespace, currentPod.Name, w, flusher)
		}()
	}
	
	// Watch for new pods (for rolling updates, restarts, etc.)
	watchOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName),
		Watch:         true,
	}
	
	watcher, err := k8sClient.Clientset.CoreV1().Pods(namespace).Watch(ctx, watchOpts)
	if err != nil {
		return fmt.Errorf("failed to create pod watcher: %v", err)
	}
	defer watcher.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				log.Printf("Watch error: %v", event.Object)
				continue
			}
			
			if event.Type == watch.Added || event.Type == watch.Modified {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					continue
				}
				
				// Only stream from running pods
				if pod.Status.Phase == corev1.PodRunning {
					utils.WriteSSEData(w, fmt.Sprintf("New pod detected: %s, switching log stream...", pod.Name))
					flusher.Flush()
					
					// Stream new pod logs
					go func(podName string) {
						s.streamPodLogs(ctx, k8sClient, namespace, podName, w, flusher)
					}(pod.Name)
				}
			}
		}
	}
}

// streamPodLogs streams logs from pod level (all containers combined)
func (s *DeploymentService) streamPodLogs(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName string, w http.ResponseWriter, flusher http.Flusher) error {
	// Wait for pod to be ready
	err := s.waitForPodReady(ctx, k8sClient, namespace, podName)
	if err != nil {
		log.Printf("Pod %s not ready: %v", podName, err)
		return err
	}
	
	// Stream logs from all containers in pod (combined)
	logOpts := &corev1.PodLogOptions{
		Follow:     true,
		Timestamps: false,
		TailLines:  int64Ptr(100),
		// No container specified = all containers combined
	}
	
	req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("error opening log stream for pod %s: %v", podName, err)
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
		return fmt.Errorf("error reading logs from pod %s: %v", podName, err)
	}
	
	return nil
}

// waitForPodReady waits for pod to be in running state
func (s *DeploymentService) waitForPodReady(ctx context.Context, k8sClient *kubernetes.Client, namespace, podName string) error {
	// Create watch for specific pod
	watchOpts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
		Watch:         true,
	}
	
	watcher, err := k8sClient.Clientset.CoreV1().Pods(namespace).Watch(ctx, watchOpts)
	if err != nil {
		return fmt.Errorf("failed to create pod status watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Wait for pod to be ready with timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer timeoutCancel()
	
	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for pod %s to be ready", podName)
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}
			
			if event.Type == watch.Added || event.Type == watch.Modified {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					continue
				}
				
				// Check if pod is running or has logs available
				if pod.Status.Phase == corev1.PodRunning ||
					pod.Status.Phase == corev1.PodSucceeded ||
					pod.Status.Phase == corev1.PodFailed {
					return nil
				}
			}
		}
	}
}