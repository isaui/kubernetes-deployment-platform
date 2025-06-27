package utils

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	resource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GetResourceName generates a consistent, immutable resource name based on service ID
// This ensures resources can be tracked even if service name changes
func GetResourceName(service models.Service) string {
	// Use full service ID (UUID) as the resource name
	// UUID is already Kubernetes-safe and immutable
	return service.ID
}
func getMainContainerName() string {
	return "app"
}
// GetResourceLabels generates consistent labels for resources
func GetResourceLabels(service models.Service) map[string]string {
	return map[string]string{
		"app":         GetResourceName(service), // Use immutable resource name
		"service-id":  service.ID,
		"service-name": service.Name, // Keep name for human readability
		"environment": service.EnvironmentID,
		"managed-by":  "pendeploy",
	}
}

// CreateKubernetesDeployment creates a Kubernetes Deployment resource for the service
func CreateKubernetesDeployment(imageURL string, service models.Service) (*appsv1.Deployment, error) {
	// Create deployment name based on immutable resource name
	deploymentName := GetResourceName(service)

	// Determine how many replicas to use
	replicas := int32(service.Replicas)
	
	// For non-static replicas, we use min replicas
	if !service.IsStaticReplica {
		// Use min replicas when HPA will be created
		replicas = int32(service.MinReplicas)
	}

	// Create labels for resources
	labels := GetResourceLabels(service)

	// Create the deployment spec
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: service.EnvironmentID, // Use environment ID as namespace
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": GetResourceName(service), // Use immutable resource name
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  getMainContainerName(),
							Image: imageURL,
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
								// Add resource requests for HPA to work properly
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

	return deployment, nil
}

// CreateKubernetesService creates a Kubernetes Service resource for the service
func CreateKubernetesService(service models.Service) (*corev1.Service, error) {
	// Create service name based on immutable resource name
	serviceName := GetResourceName(service)

	// Create labels for resources
	labels := GetResourceLabels(service)

	// Create the service spec
	k8sService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: service.EnvironmentID, // Use environment ID as namespace
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": GetResourceName(service), // Use immutable resource name
			},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(service.Port),
					TargetPort: intstr.FromInt(service.Port),
					Protocol:   corev1.ProtocolTCP,
					Name:       "http",
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	return k8sService, nil
}

// CreateKubernetesIngress creates a Kubernetes Ingress resource for the service
func CreateKubernetesIngress(service models.Service) (*networkingv1.Ingress, error) {
	// Create ingress name based on immutable resource name
	ingressName := GetResourceName(service)
	resourceName := GetResourceName(service)

	// Create labels for resources
	labels := GetResourceLabels(service)

	// Determine the hostname to use
	// For this case, CustomDomain takes precedence, but if both are set, we use both
	var hostnames []string
	if service.CustomDomain != "" {
		hostnames = append(hostnames, service.CustomDomain)
	}
	if service.Domain != "" {
		hostnames = append(hostnames, service.Domain)
	}
	if len(hostnames) == 0 {
		// No domains specified, use default domain pattern
		hostnames = append(hostnames, fmt.Sprintf("%s.%s.app.isacitra.com", service.Name, service.EnvironmentID))
	}

	// Path type for ingress
	pathTypePrefix := networkingv1.PathTypePrefix
	
	// Create the ingress spec
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: service.EnvironmentID, // Use environment ID as namespace
			Labels:    labels,
			Annotations: map[string]string{
				"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
				"traefik.ingress.kubernetes.io/router.tls":         "true",
			},
		},
		Spec: networkingv1.IngressSpec{
			// Create rules for each hostname
			Rules: []networkingv1.IngressRule{},
			// TLS configuration
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: hostnames,
				},
			},
		},
	}
	
	// Add a rule for each hostname
	for _, host := range hostnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: resourceName, // Use immutable resource name
									Port: networkingv1.ServiceBackendPort{
										Number: int32(service.Port),
									},
								},
							},
						},
					},
				},
			},
		})
	}

	return ingress, nil
}

// CreateKubernetesHPA creates a Horizontal Pod Autoscaler for the deployment
func CreateKubernetesHPA(service models.Service) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// Create HPA name based on immutable resource name
	hpaName := GetResourceName(service)
	resourceName := GetResourceName(service)

	// Create labels for the HPA
	labels := GetResourceLabels(service)

	// Set CPU utilization target (default to 70%)
	cpuUtilization := int32(70)
	// Note: If CPUUtilization field is added to Service model in the future,
	// we can use that value instead of the default

	// Convert min replicas to int32 pointer
	minReplicas := int32(service.MinReplicas)

	// Create the HPA spec
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpaName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       resourceName, // Use immutable resource name
				APIVersion: "apps/v1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: int32(service.MaxReplicas),
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &cpuUtilization,
						},
					},
				},
			},
		},
	}

	return hpa, nil
}

// DeployToKubernetesAtomically applies all Kubernetes resources atomically
// This provides a transactional approach to resource creation
func DeployToKubernetesAtomically(imageURL string, service models.Service) error {
	// Create the deployment, service, and ingress resources
	deployment, err := CreateKubernetesDeployment(imageURL, service)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %v", err)
	}

	k8sService, err := CreateKubernetesService(service)
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	ingress, err := CreateKubernetesIngress(service)
	if err != nil {
		return fmt.Errorf("failed to create ingress: %v", err)
	}
	
	// Create and store HPA if service is configured for auto-scaling (non-static replicas)
	var hpa *autoscalingv2.HorizontalPodAutoscaler
	var applyHPA bool
	
	if !service.IsStaticReplica {
		// Create the HPA resource
		hpa, err = CreateKubernetesHPA(service)
		if err != nil {
			return fmt.Errorf("failed to create HPA: %v", err)
		}
		applyHPA = true
	}

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Create context for the operations
	ctx := context.Background()
	
	// Get consistent resource names
	resourceName := GetResourceName(service)

	// Ensure the namespace exists
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Get(ctx, service.EnvironmentID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the namespace if it doesn't exist
			_, err = k8sClient.Clientset.CoreV1().Namespaces().Create(
				ctx,
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: service.EnvironmentID,
						Labels: map[string]string{
							"managed-by": "pendeploy",
						},
					},
				},
				metav1.CreateOptions{},
			)
			if err != nil {
				return fmt.Errorf("failed to create namespace: %v", err)
			}
		} else {
			return fmt.Errorf("failed to get namespace: %v", err)
		}
	}

	// Start with Deployment
	_, err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// If deployment already exists, update it
			_, err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Update(ctx, deployment, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		} else {
			return fmt.Errorf("failed to create deployment: %v", err)
		}
	}

	// Create or update the Service
	_, err = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Create(ctx, k8sService, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// If service already exists, update it
			_, err = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Update(ctx, k8sService, metav1.UpdateOptions{})
			if err != nil {
				// Rollback the deployment if service update fails
				_ = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
				return fmt.Errorf("failed to update service: %v", err)
			}
		} else {
			// Rollback the deployment if service creation fails
			_ = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
			return fmt.Errorf("failed to create service: %v", err)
		}
	}

	// Create or update the Ingress
	_, err = k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// If ingress already exists, update it
			_, err = k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Update(ctx, ingress, metav1.UpdateOptions{})
			if err != nil {
				// Rollback the deployment and service if ingress update fails
				_ = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
				_ = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
				return fmt.Errorf("failed to update ingress: %v", err)
			}
		} else {
			// Rollback the deployment and service if ingress creation fails
			_ = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
			_ = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
			return fmt.Errorf("failed to create ingress: %v", err)
		}
	}

	// Create or update the HPA if needed
	if applyHPA && hpa != nil {
		_, err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Create(ctx, hpa, metav1.CreateOptions{})
		if err != nil {
			if errors.IsAlreadyExists(err) {
				// If HPA already exists, update it
				_, err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Update(ctx, hpa, metav1.UpdateOptions{})
				if err != nil {
					// Log error but don't rollback other resources for HPA failure
					fmt.Printf("Warning: failed to update HPA: %v\n", err)
				}
			} else {
				// Log error but don't rollback other resources for HPA failure
				fmt.Printf("Warning: failed to create HPA: %v\n", err)
			}
		}
	} else if !applyHPA {
		// If service is now static replicas, delete existing HPA if it exists
		err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			// Log warning but don't fail the deployment
			fmt.Printf("Warning: failed to delete existing HPA: %v\n", err)
		}
	}

	return nil
}

// GetKubernetesResourceStatus gets the status of all resources for a service via Kubernetes API
func GetKubernetesResourceStatus(service models.Service) (map[string]interface{}, error) {
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	ctx := context.Background()
	resourceName := GetResourceName(service) // This is just service.ID
	status := make(map[string]interface{})

	// Get Deployment status
	deployment, err := k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["deployment"] = map[string]interface{}{
			"name":              deployment.Name,
			"replicas":          deployment.Status.Replicas,
			"readyReplicas":     deployment.Status.ReadyReplicas,
			"availableReplicas": deployment.Status.AvailableReplicas,
			"conditions":        deployment.Status.Conditions,
		}
	}

	// Get Service status
	svc, err := k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["service"] = map[string]interface{}{
			"name":      svc.Name,
			"clusterIP": svc.Spec.ClusterIP,
			"ports":     svc.Spec.Ports,
		}
	}

	// Get Ingress status
	ingress, err := k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["ingress"] = map[string]interface{}{
			"name":  ingress.Name,
			"hosts": ingress.Spec.Rules,
		}
	}

	// Get HPA status if exists
	hpa, err := k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["hpa"] = map[string]interface{}{
			"name":           hpa.Name,
			"currentReplicas": hpa.Status.CurrentReplicas,
			"desiredReplicas": hpa.Status.DesiredReplicas,
		}
	}

	return status, nil
}

// DeleteKubernetesResources deletes all Kubernetes resources for the service
func DeleteKubernetesResources(service models.Service) error {
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Create context for the operations
	ctx := context.Background()
	
	// Get consistent resource name
	resourceName := GetResourceName(service)

	// Delete resources in reverse order (HPA -> Ingress -> Service -> Deployment)
	// This ensures proper cleanup

	// Delete HPA if exists
	err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete HPA: %v", err)
	}

	// Delete Ingress
	err = k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Ingress: %v", err)
	}

	// Delete Service
	err = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %v", err)
	}

	// Delete Deployment
	err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment: %v", err)
	}

	return nil
}

// Helper function to convert environment variables map to Kubernetes EnvVar slice
func createEnvVarsFromMap(envVars models.EnvVars) []corev1.EnvVar {
	if len(envVars) == 0 {
		return nil
	}

	result := make([]corev1.EnvVar, 0, len(envVars))
	for key, value := range envVars {
		result = append(result, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
	
	return result
}