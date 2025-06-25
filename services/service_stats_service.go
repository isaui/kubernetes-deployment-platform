package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceStatsService handles operations related to Kubernetes services statistics
type ServiceStatsService struct{}

// NewServiceStatsService creates a new service stats service
func NewServiceStatsService() *ServiceStatsService {
	return &ServiceStatsService{}
}

// GetServiceStats returns statistics about services in the specified namespace
func (s *ServiceStatsService) GetServiceStats(namespace string) (dto.ServiceStatsResponse, error) {
	ctx := context.Background()

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.ServiceStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get services from specified namespace
	serviceList, err := kubeClient.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return dto.ServiceStatsResponse{}, fmt.Errorf("failed to list services: %v", err)
	}

	// Get endpoints to count active endpoints for each service
	endpointList, err := kubeClient.Clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// Log but continue - we can show services without endpoint counts
		fmt.Printf("Warning: Error getting endpoints: %v\n", err)
	}

	// Build endpoints lookup map
	endpointsMap := make(map[string]int)
	if endpointList != nil {
		for _, endpoint := range endpointList.Items {
			key := fmt.Sprintf("%s/%s", endpoint.Namespace, endpoint.Name)
			count := 0
			for _, subset := range endpoint.Subsets {
				count += len(subset.Addresses)
			}
			endpointsMap[key] = count
		}
	}

	// Process each service
	serviceStats := make([]dto.ServiceStats, 0, len(serviceList.Items))
	for _, service := range serviceList.Items {
		// Get endpoint count if available
		endpointCount := 0
		key := fmt.Sprintf("%s/%s", service.Namespace, service.Name)
		if count, ok := endpointsMap[key]; ok {
			endpointCount = count
		}

		// Convert selector to string array for easier frontend display
		selector := make([]string, 0, len(service.Spec.Selector))
		for k, v := range service.Spec.Selector {
			selector = append(selector, fmt.Sprintf("%s: %s", k, v))
		}

		// Calculate pod count based on selector (approximation)
		podCount := endpointCount // Use endpoint count as a reasonable approximation

		// Process ports
		ports := make([]dto.ServicePort, 0, len(service.Spec.Ports))
		for _, port := range service.Spec.Ports {
			targetPort := int32(0)
			if port.TargetPort.IntVal != 0 {
				targetPort = port.TargetPort.IntVal
			}
			
			ports = append(ports, dto.ServicePort{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.Port,
				TargetPort: targetPort,
				NodePort:   port.NodePort,
			})
		}

		// Get load balancer info
		loadBalancer := ""
		if service.Spec.Type == "LoadBalancer" && len(service.Status.LoadBalancer.Ingress) > 0 {
			if service.Status.LoadBalancer.Ingress[0].IP != "" {
				loadBalancer = service.Status.LoadBalancer.Ingress[0].IP
			} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
				loadBalancer = service.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		// Add service stats
		serviceStats = append(serviceStats, dto.ServiceStats{
			Name:          service.Name,
			Namespace:     service.Namespace,
			Type:          string(service.Spec.Type),
			ClusterIP:     service.Spec.ClusterIP,
			ExternalIPs:   service.Spec.ExternalIPs,
			LoadBalancer:  loadBalancer,
			Ports:         ports,
			Selector:      selector,
			PodCount:      podCount,
			EndpointCount: endpointCount,
			Created:       service.CreationTimestamp.Format(time.RFC3339),
		})
	}

	return dto.ServiceStatsResponse{
		Services: serviceStats,
	}, nil
}
