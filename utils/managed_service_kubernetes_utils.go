// utils/managed_kubernetes_utils.go
package utils

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// Allocate NodePort for primary TCP service
	nodePort, err := AllocateNodePort(service.ManagedType, serverIP)
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

	// Deploy all services (NodePort for TCP, ClusterIP for HTTP)
	if err := deployAllManagedServices(ctx, k8sClient, service, nodePort); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("services: %v", err))
	}

	// Deploy ingresses only for HTTP services
	if err := deployManagedIngresses(ctx, k8sClient, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("ingresses: %v", err))
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
							Image: containerImage,
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
			Replicas:            &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": resourceName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "managed-service",
							Image: containerImage,
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
		_, err = client.Clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
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
		log.Printf("PVC %s already exists, skipping creation", pvc.Name)
		return nil
	}
	return err
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