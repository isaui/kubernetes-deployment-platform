// utils/managed_kubernetes_utils.go
package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// getOrAllocateNodePort gets existing NodePort from service or allocates new one
func getOrAllocateNodePort(ctx context.Context, client *kubernetes.Client, service models.Service, serverIP string) (int, error) {
	resourceName := GetResourceName(service)

	// Try to get existing service to reuse NodePort
	existingService, err := client.Clientset.CoreV1().Services(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		// Service exists, check if it has NodePort
		for _, port := range existingService.Spec.Ports {
			if port.NodePort != 0 {
				log.Printf("Reusing existing NodePort %d for service %s", port.NodePort, service.Name)
				return int(port.NodePort), nil
			}
		}
	}

	// Service doesn't exist or doesn't have NodePort, allocate new one
	return AllocateNodePort(service.ManagedType, serverIP)
}

// DeployManagedServiceToKubernetes deploys a managed service to Kubernetes with NodePort for TCP services
func DeployManagedServiceToKubernetes(service models.Service, serverIP string) (*models.Service, error) {
	service.Status = "building"

	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		service.Status = "failed"
		return &service, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	ctx := context.Background()

	if err := EnsureNamespaceExists(service.EnvironmentID); err != nil {
		service.Status = "failed"
		return &service, fmt.Errorf("failed to ensure namespace: %v", err)
	}

	// Check if service already exists to reuse NodePort
	nodePort, err := getOrAllocateNodePort(ctx, k8sClient, service, serverIP)
	if err != nil {
		service.Status = "failed"
		return &service, fmt.Errorf("failed to allocate NodePort: %v", err)
	}

	// Set port and env vars using enhanced utils with NodePort
	service.Port = GetManagedServicePort(service.ManagedType)
	service.EnvVars = GenerateManagedServiceEnvVars(service, serverIP, nodePort)

	var deploymentErrors []string

	// Deploy workload (StatefulSet/Deployment)
	serviceType := GetManagedServiceType(service.ManagedType)
	if serviceType == "StatefulSet" {
		if err := deployStatefulSet(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("statefulset: %v", err))
		}
	} else {
		if err := deployManagedDeployment(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("deployment: %v", err))
		}
		if RequiresPersistentStorage(service.ManagedType) {
			if err := createManagedServicePVC(ctx, k8sClient, service); err != nil {
				deploymentErrors = append(deploymentErrors, fmt.Sprintf("pvc: %v", err))
			}
		}
	}

	// Check if services already exist - if yes, skip service/ingress deployment
	// This prevents NodePort conflicts during StatefulSet resource updates
	resourceName := GetResourceName(service)
	_, serviceErr := k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if serviceErr != nil && errors.IsNotFound(serviceErr) {
		// Services don't exist, deploy them
		log.Printf("Deploying new services and ingresses for %s", service.Name)

		// Deploy all services (NodePort for TCP, ClusterIP for HTTP)
		if err := deployAllManagedServices(ctx, k8sClient, service, nodePort); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("services: %v", err))
		}

		// Deploy ingresses only for HTTP services
		if err := deployManagedIngresses(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("ingresses: %v", err))
		}
	} else {
		log.Printf("Skipping service/ingress deployment - resources already exist for %s", service.Name)
	}

	if len(deploymentErrors) > 0 {
		service.Status = "failed"
		return &service, fmt.Errorf("deployment failed: %s", strings.Join(deploymentErrors, "; "))
	}

	// Set external connection info
	service.Domain = fmt.Sprintf("%s:%d", serverIP, nodePort) // For primary TCP service
	service.Status = "running"

	log.Printf("Successfully deployed managed service: %s (%s) with NodePort %d", service.Name, service.ManagedType, nodePort)
	return &service, nil
}

// deployAllManagedServices creates all required services with appropriate exposure
func deployAllManagedServices(ctx context.Context, client *kubernetes.Client, service models.Service, primaryNodePort int) error {
	serviceConfigs := GetManagedServiceExposureConfig(service.ManagedType)

	for _, config := range serviceConfigs {
		var k8sService *corev1.Service

		if config.ExposureType == "NodePort" && config.Name == "primary" {
			// Primary TCP service gets allocated NodePort
			k8sService = createNodePortServiceSpec(service, config, primaryNodePort)
		} else if config.ExposureType == "NodePort" {
			// Secondary TCP services would need separate port allocation (future)
			k8sService = createClusterIPServiceSpec(service, config)
		} else {
			// HTTP services get ClusterIP for internal access (exposed via Ingress)
			k8sService = createClusterIPServiceSpec(service, config)
		}

		if err := applyManagedService(ctx, client, k8sService); err != nil {
			return fmt.Errorf("service %s: %v", config.Name, err)
		}
	}
	return nil
}

// deployManagedIngresses creates ingresses only for HTTP services
func deployManagedIngresses(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	serviceConfigs := GetManagedServiceExposureConfig(service.ManagedType)

	for _, config := range serviceConfigs {
		if config.IsHTTP && config.ExposureType == "Ingress" {
			// Create HTTP Ingress for web services (MinIO console, RabbitMQ management)
			ingress := createManagedIngressSpec(service, config)
			if err := applyManagedIngress(ctx, client, ingress); err != nil {
				return fmt.Errorf("http ingress %s: %v", config.Name, err)
			}
			log.Printf("Created HTTP Ingress for %s (%s)", service.Name, config.Name)
		}
	}
	return nil
}

// createNodePortServiceSpec creates NodePort Service for TCP services
func createNodePortServiceSpec(service models.Service, config ServiceExposureConfig, nodePort int) *corev1.Service {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	serviceName := resourceName

	// Add suffix for secondary services
	if config.Name != "primary" {
		serviceName = fmt.Sprintf("%s-%s", resourceName, config.Name)
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{"app": resourceName},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(config.Port),
					TargetPort: intstr.FromInt(config.Port),
					Protocol:   corev1.ProtocolTCP,
					Name:       config.Name,
					NodePort:   int32(nodePort),
				},
			},
		},
	}
}

// createClusterIPServiceSpec creates ClusterIP Service for internal/HTTP services
func createClusterIPServiceSpec(service models.Service, config ServiceExposureConfig) *corev1.Service {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	serviceName := resourceName

	// Add suffix for secondary services
	if config.Name != "primary" {
		serviceName = fmt.Sprintf("%s-%s", resourceName, config.Name)
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": resourceName},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(config.Port),
					TargetPort: intstr.FromInt(config.Port),
					Protocol:   corev1.ProtocolTCP,
					Name:       config.Name,
				},
			},
		},
	}
}

// createStatefulSetSpec creates StatefulSet with all required ports
func createStatefulSetSpec(service models.Service) *appsv1.StatefulSet {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	replicas := int32(1)
	containerImage := getManagedServiceImage(service.ManagedType, service.Version)

	// Get all ports for this service type
	exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)
	var containerPorts []corev1.ContainerPort
	for _, config := range exposureConfigs {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: int32(config.Port),
			Protocol:      corev1.ProtocolTCP,
			Name:          config.Name,
		})
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: resourceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": resourceName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "managed-service",
							Image:   containerImage,
						Command: getManagedServiceCommand(service.ManagedType),
						Args:    getManagedServiceArgs(service.ManagedType),
							Ports: containerPorts,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(service.CPULimit),
									corev1.ResourceMemory: resource.MustParse(service.MemoryLimit),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
							Env: createEnvVarsFromMap(service.EnvVars),
						},
					},
				},
			},
		},
	}

	// Add storage if required
	if RequiresPersistentStorage(service.ManagedType) {
		statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: getManagedServiceDataPath(service.ManagedType),
			},
		}

		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "data",
					Labels: labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(service.StorageSize),
						},
					},
				},
			},
		}
	}

	return statefulSet
}

// createManagedDeploymentSpec creates Deployment with all required ports
func createManagedDeploymentSpec(service models.Service) *appsv1.Deployment {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)

	replicas := int32(service.Replicas)
	if !service.IsStaticReplica {
		replicas = int32(service.MinReplicas)
	}

	containerImage := getManagedServiceImage(service.ManagedType, service.Version)

	// Get all ports for this service type
	exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)
	var containerPorts []corev1.ContainerPort
	for _, config := range exposureConfigs {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: int32(config.Port),
			Protocol:      corev1.ProtocolTCP,
			Name:          config.Name,
		})
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: int32Ptr(1),
			Replicas:             &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": resourceName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "managed-service",
							Image:   containerImage,
						Command: getManagedServiceCommand(service.ManagedType),
						Args:    getManagedServiceArgs(service.ManagedType),
							Ports: containerPorts,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(service.CPULimit),
									corev1.ResourceMemory: resource.MustParse(service.MemoryLimit),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
							Env: createEnvVarsFromMap(service.EnvVars),
						},
					},
				},
			},
		},
	}

	// Add storage if required
	if RequiresPersistentStorage(service.ManagedType) {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: getManagedServiceDataPath(service.ManagedType),
			},
		}

		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-data", resourceName),
					},
				},
			},
		}
	}

	return deployment
}

// createManagedIngressSpec creates Ingress specification for HTTP services only
func createManagedIngressSpec(service models.Service, config ServiceExposureConfig) *networkingv1.Ingress {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	ingressName := resourceName
	serviceName := resourceName

	// Add suffix for secondary ingresses
	if config.Name != "primary" {
		ingressName = fmt.Sprintf("%s-%s", resourceName, config.Name)
		serviceName = fmt.Sprintf("%s-%s", resourceName, config.Name)
	}

	hostname := GetManagedServiceExternalDomain(service, config.Name)
	pathTypePrefix := networkingv1.PathTypePrefix
	tlsSecretName := fmt.Sprintf("%s-tls", ingressName)

	// HTTP Ingress annotations
	annotations := map[string]string{
		"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
		"traefik.ingress.kubernetes.io/router.tls":         "true",
		"cert-manager.io/cluster-issuer":                   "letsencrypt-prod",
	}

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Namespace:   service.EnvironmentID,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: int32(config.Port),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{hostname},
					SecretName: tlsSecretName,
				},
			},
		},
	}
}

// createManagedServicePVC creates PVC for Deployment-based services
func createManagedServicePVC(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	if !RequiresPersistentStorage(service.ManagedType) {
		return nil
	}

	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-data", resourceName),
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(service.StorageSize),
				},
			},
		},
	}

	return applyPVC(ctx, client, pvc)
}

// Helper functions for StatefulSet and Deployment deployment
func deployStatefulSet(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	statefulSet := createStatefulSetSpec(service)
	return applyStatefulSet(ctx, client, statefulSet)
}

func deployManagedDeployment(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	deployment := createManagedDeploymentSpec(service)
	return applyManagedDeployment(ctx, client, deployment)
}

// Apply functions
func applyStatefulSet(ctx context.Context, client *kubernetes.Client, statefulSet *appsv1.StatefulSet) error {
	_, err := client.Clientset.AppsV1().StatefulSets(statefulSet.Namespace).Create(ctx, statefulSet, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		// For StatefulSet, resource changes require scale-down-update-scale-up
		if err := updateStatefulSetWithScaling(ctx, client, statefulSet); err != nil {
			return err
		}
		// Update successful, return nil
		return nil
	}
	return err
}

func applyManagedDeployment(ctx context.Context, client *kubernetes.Client, deployment *appsv1.Deployment) error {
	_, err := client.Clientset.AppsV1().Deployments(deployment.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	}
	return err
}

func applyManagedService(ctx context.Context, client *kubernetes.Client, service *corev1.Service) error {
	_, err := client.Clientset.CoreV1().Services(service.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.CoreV1().Services(service.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	}
	return err
}

func applyManagedIngress(ctx context.Context, client *kubernetes.Client, ingress *networkingv1.Ingress) error {
	_, err := client.Clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	}
	return err
}

func applyPVC(ctx context.Context, client *kubernetes.Client, pvc *corev1.PersistentVolumeClaim) error {
	_, err := client.Clientset.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		// Check if we need to expand storage
		existingPVC, getErr := client.Clientset.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(ctx, pvc.Name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		requested := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		existing := existingPVC.Spec.Resources.Requests[corev1.ResourceStorage]

		if requested.Cmp(existing) > 0 {
			log.Printf("Expanding PVC %s from %s to %s", pvc.Name, existing.String(), requested.String())
			existingPVC.Spec.Resources.Requests[corev1.ResourceStorage] = requested
			_, err = client.Clientset.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(ctx, existingPVC, metav1.UpdateOptions{})
			return err
		}

		return nil
	}
	return err
}

// updateStatefulSetWithScaling updates StatefulSet by scaling down, updating spec, then scaling up
func updateStatefulSetWithScaling(ctx context.Context, client *kubernetes.Client, newStatefulSet *appsv1.StatefulSet) error {
	log.Printf("Updating StatefulSet %s via scale-down-update-scale-up", newStatefulSet.Name)

	// Step 1: Scale down to 0 (get fresh object first)
	existingStatefulSet, err := client.Clientset.AppsV1().StatefulSets(newStatefulSet.Namespace).Get(ctx, newStatefulSet.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get existing StatefulSet: %v", err)
	}

	zeroReplicas := int32(0)
	existingStatefulSet.Spec.Replicas = &zeroReplicas
	_, err = client.Clientset.AppsV1().StatefulSets(newStatefulSet.Namespace).Update(ctx, existingStatefulSet, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale down StatefulSet: %v", err)
	}
	log.Printf("Scaled down StatefulSet %s to 0 replicas", newStatefulSet.Name)

	// Step 2: Wait for pods to terminate
	time.Sleep(5 * time.Second)

	// Step 3: Get fresh object again and update template spec + scale up
	existingStatefulSet, err = client.Clientset.AppsV1().StatefulSets(newStatefulSet.Namespace).Get(ctx, newStatefulSet.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get StatefulSet for template update: %v", err)
	}

	// Update template spec with new resource limits
	existingStatefulSet.Spec.Template = newStatefulSet.Spec.Template
	existingStatefulSet.Spec.VolumeClaimTemplates = newStatefulSet.Spec.VolumeClaimTemplates
	
	// Scale back up to 1
	oneReplica := int32(1)
	existingStatefulSet.Spec.Replicas = &oneReplica
	_, err = client.Clientset.AppsV1().StatefulSets(newStatefulSet.Namespace).Update(ctx, existingStatefulSet, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale up StatefulSet: %v", err)
	}

	log.Printf("Successfully updated StatefulSet %s via scaling", newStatefulSet.Name)
	return nil
}

// Service helper functions
func getManagedServiceImage(managedType, version string) string {
	images := map[string]string{
		"postgresql": fmt.Sprintf("postgres:%s", version),
		"mysql":      fmt.Sprintf("mysql:%s", version),
		"redis":      fmt.Sprintf("redis:%s", version),
		"mongodb":    fmt.Sprintf("mongo:%s", version),
		"minio":      fmt.Sprintf("minio/minio:%s", version),
		"rabbitmq":   fmt.Sprintf("rabbitmq:%s-management", version),
	}

	if image, exists := images[managedType]; exists {
		return image
	}
	return "alpine:latest"
}

func getManagedServiceDataPath(managedType string) string {
	paths := map[string]string{
		"postgresql": "/var/lib/postgresql/data",
		"mysql":      "/var/lib/mysql",
		"redis":      "/data",
		"mongodb":    "/data/db",
		"minio":      "/data",
		"rabbitmq":   "/var/lib/rabbitmq",
	}

	if path, exists := paths[managedType]; exists {
		return path
	}
	return "/data"
}

// getManagedServiceCommand returns the command for container startup
// Returns nil to use image default ENTRYPOINT for most services
func getManagedServiceCommand(managedType string) []string {
	switch managedType {
	case "minio":
		// MinIO needs explicit command
		return []string{"minio"}
	// Other services use image defaults:
	// - postgres: uses default ENTRYPOINT ["docker-entrypoint.sh"]
	// - mysql: uses default ENTRYPOINT ["docker-entrypoint.sh"]
	// - redis: uses default ENTRYPOINT ["docker-entrypoint.sh"]
	// - mongo: uses default ENTRYPOINT ["docker-entrypoint.sh"]
	// - rabbitmq: uses default ENTRYPOINT ["/opt/rabbitmq/sbin/rabbitmq-server"]
	default:
		return nil // Use default image ENTRYPOINT
	}
}

// getManagedServiceArgs returns the args for container startup
// Returns nil to use image default CMD for most services
func getManagedServiceArgs(managedType string) []string {
	switch managedType {
	case "minio":
		// MinIO needs explicit server command with console
		return []string{"server", "/data", "--console-address", ":9001"}
	// Other services use image defaults:
	// - postgres: uses default CMD ["postgres"]
	// - mysql: uses default CMD ["mysqld"]
	// - redis: uses default CMD ["redis-server"]
	// - mongo: uses default CMD ["mongod"]
	// - rabbitmq: uses default args (none needed)
	default:
		return nil // Use default image CMD
	}
}
