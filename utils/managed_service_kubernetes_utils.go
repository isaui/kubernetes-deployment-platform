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

// DeployManagedServiceToKubernetes deploys a managed service to Kubernetes
func DeployManagedServiceToKubernetes(service models.Service) (*models.Service, error) {
	// Update service status to building
	service.Status = "building"
	
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		service.Status = "failed"
		return &service, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	ctx := context.Background()
	
	// Ensure namespace exists
	if err := EnsureNamespaceExists(service.EnvironmentID); err != nil {
		service.Status = "failed"
		return &service, fmt.Errorf("failed to ensure namespace: %v", err)
	}

	// Set port based on managed type
	service.Port = GetManagedServicePort(service.ManagedType)
	
	// Generate secure environment variables
	service.EnvVars = GenerateManagedServiceEnvVars(service)
	
	var deploymentErrors []string

	// Deploy core resources based on service type
	serviceType := GetManagedServiceType(service.ManagedType)
	if serviceType == "StatefulSet" {
		if err := deployStatefulSet(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("statefulset: %v", err))
		}
	} else {
		if err := deployManagedDeployment(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("deployment: %v", err))
		}
	}

	// Deploy Service for internal access
	if err := deployManagedService(ctx, k8sClient, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("service: %v", err))
	}

	// Deploy Ingress for external access
	if err := deployManagedIngress(ctx, k8sClient, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("ingress: %v", err))
	}

	// Create PVC if storage is required
	if RequiresPersistentStorage(service.ManagedType) {
		if err := createManagedServicePVC(ctx, k8sClient, service); err != nil {
			deploymentErrors = append(deploymentErrors, fmt.Sprintf("pvc: %v", err))
		}
	}

	// Update service status based on deployment result
	if len(deploymentErrors) > 0 {
		service.Status = "failed"
		return &service, fmt.Errorf("deployment failed: %s", strings.Join(deploymentErrors, "; "))
	}

	// Set external domain
	service.Domain = GetManagedServiceExternalDomain(service)
	service.Status = "running"
	
	log.Printf("Successfully deployed managed service: %s (%s)", service.Name, service.ManagedType)
	return &service, nil
}

// deployStatefulSet creates a StatefulSet for database services
func deployStatefulSet(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	statefulSet := createStatefulSetSpec(service)
	return applyStatefulSet(ctx, client, statefulSet)
}

// deployManagedDeployment creates a Deployment for stateless managed services
func deployManagedDeployment(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	deployment := createManagedDeploymentSpec(service)
	return applyManagedDeployment(ctx, client, deployment)
}

// deployManagedService creates a Service for managed service
func deployManagedService(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	k8sService := createManagedServiceSpec(service)
	return applyManagedService(ctx, client, k8sService)
}

// deployManagedIngress creates an Ingress for external access
func deployManagedIngress(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	ingress := createManagedIngressSpec(service)
	return applyManagedIngress(ctx, client, ingress)
}

// createStatefulSetSpec creates StatefulSet specification
func createStatefulSetSpec(service models.Service) *appsv1.StatefulSet {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	
	replicas := int32(1) // Managed services always single replica for data consistency
	
	// Get container image based on managed type
	containerImage := getManagedServiceImage(service.ManagedType, service.Version)

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
				MatchLabels: map[string]string{
					"app": resourceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "managed-service",
							Image: containerImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(service.Port),
									Protocol:      corev1.ProtocolTCP,
								},
							},
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

	// Add volume mounts and volume claim templates if storage is required
	if RequiresPersistentStorage(service.ManagedType) {
		// Add volume mount
		statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: getManagedServiceDataPath(service.ManagedType),
			},
		}

		// Add volume claim template
		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "data",
					Labels: labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
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

// createManagedDeploymentSpec creates Deployment specification for stateless services
func createManagedDeploymentSpec(service models.Service) *appsv1.Deployment {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	
	replicas := int32(service.Replicas)
	if !service.IsStaticReplica {
		replicas = int32(service.MinReplicas)
	}

	containerImage := getManagedServiceImage(service.ManagedType, service.Version)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: int32Ptr(1),
			Replicas:            &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": resourceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "managed-service",
							Image: containerImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(service.Port),
									Protocol:      corev1.ProtocolTCP,
								},
							},
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
}

// createManagedServiceSpec creates Service specification
func createManagedServiceSpec(service models.Service) *corev1.Service {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": resourceName,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(service.Port),
					TargetPort: intstr.FromInt(service.Port),
					Protocol:   corev1.ProtocolTCP,
					Name:       "service",
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// createManagedIngressSpec creates Ingress specification for external access
func createManagedIngressSpec(service models.Service) *networkingv1.Ingress {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	hostname := GetManagedServiceExternalDomain(service)
	pathTypePrefix := networkingv1.PathTypePrefix

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
			Annotations: map[string]string{
				"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
				"traefik.ingress.kubernetes.io/router.tls":         "true",
				// TCP mode for databases
				"traefik.ingress.kubernetes.io/service.serversscheme": "tcp",
			},
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
											Name: resourceName,
											Port: networkingv1.ServiceBackendPort{
												Number: int32(service.Port),
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
					Hosts: []string{hostname},
				},
			},
		},
	}
}

// createManagedServicePVC creates PVC for managed services that need storage
func createManagedServicePVC(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	if !RequiresPersistentStorage(service.ManagedType) {
		return nil // Skip if no storage needed
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
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(service.StorageSize),
				},
			},
		},
	}

	return applyPVC(ctx, client, pvc)
}

// Helper functions for applying K8s resources
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
		// PVC cannot be updated, just return success if exists
		log.Printf("PVC %s already exists, skipping creation", pvc.Name)
		return nil
	}
	return err
}

// getManagedServiceImage returns Docker image for managed service
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
	return "alpine:latest" // fallback
}

// getManagedServiceDataPath returns data mount path for each service type
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
	return "/data" // fallback
}