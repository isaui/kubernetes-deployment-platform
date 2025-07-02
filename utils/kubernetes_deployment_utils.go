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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeployToKubernetesAtomically deploys all Kubernetes resources with idempotent approach
// Returns updated service with deployment status
func DeployToKubernetesAtomically(imageURL string, service models.Service) (*models.Service, error) {
	// Update service status to building
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

	var deploymentErrors []string

	// Deploy core resources
	if err := deployDeployment(ctx, k8sClient, imageURL, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("deployment: %v", err))
	}

	if err := deployService(ctx, k8sClient, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("service: %v", err))
	}

	if err := deployIngress(ctx, k8sClient, service); err != nil {
		deploymentErrors = append(deploymentErrors, fmt.Sprintf("ingress: %v", err))
	}

	// Handle HPA based on scaling configuration
	if err := handleHPA(ctx, k8sClient, service); err != nil {
		log.Printf("Warning - HPA operation failed: %v", err)
	}

	// Update service status based on deployment result
	if len(deploymentErrors) > 0 {
		service.Status = "failed"
		return &service, fmt.Errorf("deployment failed: %s", strings.Join(deploymentErrors, "; "))
	}

	// Set domain if not already set
	if service.Domain == "" {
		service.Domain = GetDefaultDomainName(service)
	}

	service.Status = "running"
	service.UpdatedAt = time.Now()
	
	log.Printf("Successfully deployed service: %s", GetResourceName(service))
	return &service, nil
}

// Core deployment functions

func deployDeployment(ctx context.Context, client *kubernetes.Client, imageURL string, service models.Service) error {
	deployment := createDeploymentSpec(imageURL, service)
	return applyDeployment(ctx, client, deployment)
}

func deployService(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	k8sService := createServiceSpec(service)
	return applyService(ctx, client, k8sService)
}

func deployIngress(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	ingress := createIngressSpec(service)
	return applyIngress(ctx, client, ingress)
}

func handleHPA(ctx context.Context, client *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	if service.IsStaticReplica {
		return deleteHPA(ctx, client, service.EnvironmentID, resourceName)
	}

	hpa := createHPASpec(service)
	return applyHPA(ctx, client, hpa)
}

// Kubernetes apply functions

func applyDeployment(ctx context.Context, client *kubernetes.Client, deployment *appsv1.Deployment) error {
	_, err := client.Clientset.AppsV1().Deployments(deployment.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	}
	return err
}

func applyService(ctx context.Context, client *kubernetes.Client, service *corev1.Service) error {
	_, err := client.Clientset.CoreV1().Services(service.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.CoreV1().Services(service.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	}
	return err
}

func applyIngress(ctx context.Context, client *kubernetes.Client, ingress *networkingv1.Ingress) error {
	_, err := client.Clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	}
	return err
}

func applyHPA(ctx context.Context, client *kubernetes.Client, hpa *autoscalingv2.HorizontalPodAutoscaler) error {
	_, err := client.Clientset.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Create(ctx, hpa, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		_, err = client.Clientset.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Update(ctx, hpa, metav1.UpdateOptions{})
	}
	return err
}

func deleteHPA(ctx context.Context, client *kubernetes.Client, namespace, resourceName string) error {
	err := client.Clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		log.Printf("HPA %s not found, nothing to delete", resourceName)
		return nil
	}
	if err != nil {
		log.Printf("Warning: failed to delete HPA %s: %v", resourceName, err)
		return err
	}
	log.Printf("HPA %s deleted successfully", resourceName)
	return nil
}

// Resource specification builders

func createDeploymentSpec(imageURL string, service models.Service) *appsv1.Deployment {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	
	replicas := int32(service.Replicas)
	if !service.IsStaticReplica {
		replicas = int32(service.MinReplicas)
	}

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

func createServiceSpec(service models.Service) *corev1.Service {
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
					Name:       "http",
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func createIngressSpec(service models.Service) *networkingv1.Ingress {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	hostnames := buildHostnames(service)
	pathTypePrefix := networkingv1.PathTypePrefix
	
	// Generate TLS secret name based on service
	// Option 1: Standard approach (recommended)
	tlsSecretName := fmt.Sprintf("%s-tls", resourceName)
	
	// Option 2: Replace hyphens if you're paranoid (NOT needed)
	// tlsSecretName := strings.ReplaceAll(resourceName, "-", "") + "tls"

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
			Annotations: map[string]string{
				// Traefik configuration
				"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
				"traefik.ingress.kubernetes.io/router.tls": "true",
				
				// Cert-manager configuration
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
				
				// Optional: HTTP to HTTPS redirect (Traefik handles this automatically for websecure)
				// "traefik.ingress.kubernetes.io/redirect-permanent": "true",
				// "traefik.ingress.kubernetes.io/redirect-scheme": "https",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      hostnames,
					SecretName: tlsSecretName, // ✅ This is the key fix!
				},
			},
		},
	}

	// Add rules for each hostname
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
		})
	}

	return ingress
}

func createHPASpec(service models.Service) *autoscalingv2.HorizontalPodAutoscaler {
	resourceName := GetResourceName(service)
	labels := GetResourceLabels(service)
	minReplicas := int32(service.MinReplicas)
	cpuUtilization := int32(70)

	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: service.EnvironmentID,
			Labels:    labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       resourceName,
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
}

// Helper functions

func buildHostnames(service models.Service) []string {
	var hostnames []string
	
	if service.CustomDomain != "" {
		hostnames = append(hostnames, service.CustomDomain)
	}
	if service.Domain != "" {
		hostnames = append(hostnames, service.Domain)
	}
	if len(hostnames) == 0 {
		hostnames = append(hostnames, GetDefaultDomainName(service))
	}
	
	return hostnames
}