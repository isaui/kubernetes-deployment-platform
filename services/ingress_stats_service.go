package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressStatsService handles operations related to Kubernetes ingress resources statistics
type IngressStatsService struct{}

// NewIngressStatsService creates a new ingress stats service
func NewIngressStatsService() *IngressStatsService {
	return &IngressStatsService{}
}

// GetIngressStats returns statistics about ingress resources in the specified namespace
func (s *IngressStatsService) GetIngressStats(namespace string) (dto.IngressStatsResponse, error) {
	ctx := context.Background()

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.IngressStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get ingress resources
	ingressList, err := kubeClient.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return dto.IngressStatsResponse{}, fmt.Errorf("failed to list ingresses: %v", err)
	}

	// Process each ingress
	ingressStats := make([]dto.IngressStats, 0, len(ingressList.Items))
	for _, ingress := range ingressList.Items {
		// Extract rules
		rules := make([]dto.IngressRule, 0)
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				// Extract service name and port
				serviceName := path.Backend.Service.Name
				servicePort := int32(0)
				
				// Handle different port types (number or name)
				if path.Backend.Service.Port.Number > 0 {
					servicePort = path.Backend.Service.Port.Number
				}

				// Create IngressRule
				pathType := ""
				if path.PathType != nil {
					pathType = string(*path.PathType)
				}
				
				rules = append(rules, dto.IngressRule{
					Host:     rule.Host,
					Path:     path.Path,
					PathType: pathType,
					Service:  serviceName,
					Port:     servicePort,
				})
			}
		}

		// Extract TLS configurations
		tlsConfigs := make([]dto.IngressTLS, 0)
		for _, tls := range ingress.Spec.TLS {
			tlsConfigs = append(tlsConfigs, dto.IngressTLS{
				Hosts:      tls.Hosts,
				SecretName: tls.SecretName,
			})
		}

		// Extract ingress class
		ingressClass := ""
		if ingress.Spec.IngressClassName != nil {
			ingressClass = *ingress.Spec.IngressClassName
		}

		// Extract LB address if available
		address := ""
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].IP
			} else if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		// Extract notable annotations
		annotations := make([]string, 0)
		for k, v := range ingress.Annotations {
			annotations = append(annotations, fmt.Sprintf("%s: %s", k, v))
		}

		// Add ingress stats
		ingressStats = append(ingressStats, dto.IngressStats{
			Name:        ingress.Name,
			Namespace:   ingress.Namespace,
			Class:       ingressClass,
			Rules:       rules,
			TLS:         tlsConfigs,
			Address:     address,
			Created:     ingress.CreationTimestamp.Format(time.RFC3339),
			Annotations: annotations,
		})
	}

	return dto.IngressStatsResponse{
		Ingresses: ingressStats,
	}, nil
}
