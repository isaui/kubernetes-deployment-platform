package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)



// CreateKubernetesDeployment creates a Kubernetes Deployment resource for the service
func CreateKubernetesDeployment(imageURL string, service models.Service) (*appsv1.Deployment, error) {
	// Create deployment name based on immutable resource name
	log.Println("Creating deployment name based on immutable resource name")
	deploymentName := GetResourceName(service)

	// Determine how many replicas to use
	replicas := int32(service.Replicas)
	log.Println("Determining how many replicas to use")
	// For non-static replicas, we use min replicas
	if !service.IsStaticReplica {
		log.Println("Using min replicas when HPA will be created")
		// Use min replicas when HPA will be created
		replicas = int32(service.MinReplicas)
	}
	log.Println("Replicas determined successfully")

	// Create labels for resources
	log.Println("Creating labels for resources")
	labels := GetResourceLabels(service)

	// Create the deployment spec
	log.Println("Creating deployment spec")
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
	log.Println("Deployment spec created successfully")

	return deployment, nil
}

// CreateKubernetesService creates a Kubernetes Service resource for the service
func CreateKubernetesService(service models.Service) (*corev1.Service, error) {
	// Create service name based on immutable resource name
	serviceName := GetResourceName(service)

	// Create labels for resources
	log.Println("Creating labels for resources")
	labels := GetResourceLabels(service)

	// Create the service spec
	log.Println("Creating service spec")
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
	log.Println("Service spec created successfully")

	return k8sService, nil
}

// CreateKubernetesIngress creates a Kubernetes Ingress resource for the service
func CreateKubernetesIngress(service models.Service) (*networkingv1.Ingress, error) {
	// Create ingress name based on immutable resource name
	log.Println("Creating ingress name based on immutable resource name")
	ingressName := GetResourceName(service)
	resourceName := GetResourceName(service)

	// Create labels for resources
	log.Println("Creating labels for resources")
	labels := GetResourceLabels(service)
    log.Println("Labels created successfully", labels)
	// Determine the hostname to use
	log.Println("Determining the hostname to use")
	// For this case, CustomDomain takes precedence, but if both are set, we use both
	var hostnames []string
    log.Println("Hostnames determined successfully", hostnames)
	if service.CustomDomain != "" {
		hostnames = append(hostnames, service.CustomDomain)
	}
	if service.Domain != "" {
		hostnames = append(hostnames, service.Domain)
	}
	if len(hostnames) == 0 {
		// No domains specified, use repo-based domain pattern
		hostnames = append(hostnames, GetDefaultDomainName(service))
	}
    log.Println("Hostnames determined successfully", hostnames)
	// Path type for ingress
	pathTypePrefix := networkingv1.PathTypePrefix
	log.Println("Path type for ingress determined successfully", pathTypePrefix)
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
	log.Println("Ingress spec created successfully")
	
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
    log.Println("Ingress rules added successfully")
	return ingress, nil
}

// CreateKubernetesHPA creates a Horizontal Pod Autoscaler for the deployment
func CreateKubernetesHPA(service models.Service) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// Create HPA name based on immutable resource name
	log.Println("Creating HPA name based on immutable resource name")
	hpaName := GetResourceName(service)
	log.Println("HPA name created successfully", hpaName)
	resourceName := GetResourceName(service)
	log.Println("Resource name created successfully", resourceName)

	// Create labels for the HPA
	log.Println("Creating labels for the HPA")
	labels := GetResourceLabels(service)
	log.Println("Labels created successfully", labels)

	// Set CPU utilization target (default to 70%)
	cpuUtilization := int32(70)
	// Note: If CPUUtilization field is added to Service model in the future,
	// we can use that value instead of the default
    log.Println("CPU utilization target set successfully", cpuUtilization)
	// Convert min replicas to int32 pointer
	minReplicas := int32(service.MinReplicas)
    log.Println("Min replicas converted to int32 pointer successfully", minReplicas)

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
    log.Println("HPA spec created successfully")
	return hpa, nil
}

// DeployToKubernetesAtomically applies all Kubernetes resources atomically
// This provides a transactional approach to resource creation
func DeployToKubernetesAtomically(imageURL string, service models.Service) error {
	// Create the deployment, service, and ingress resources
    log.Println("Creating deployment, service, and ingress resources")
	deployment, err := CreateKubernetesDeployment(imageURL, service)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %v", err)
	}
    log.Println("Deployment created successfully")

	k8sService, err := CreateKubernetesService(service)
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}
    log.Println("Service created successfully")

	ingress, err := CreateKubernetesIngress(service)
	if err != nil {
		return fmt.Errorf("failed to create ingress: %v", err)
	}
    log.Println("Ingress created successfully")
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
    log.Println("HPA created successfully")

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
    log.Println("Kubernetes client created successfully")
	// Create context for the operations
	ctx := context.Background()
    log.Println("Context created successfully")
	// Get consistent resource names
	resourceName := GetResourceName(service)
    log.Println("Resource name created successfully", resourceName)

	// Ensure the namespace exists
	err = EnsureNamespaceExists(service.EnvironmentID)
	if err != nil {
		return fmt.Errorf("failed to ensure namespace: %v", err)
	}
    log.Println("Namespace confirmed:", service.EnvironmentID)

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
    log.Println("Deployment created successfully")
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
    log.Println("Service created successfully")
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
    log.Println("Ingress created successfully")
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
    log.Println("HPA created successfully")
	return nil
}
