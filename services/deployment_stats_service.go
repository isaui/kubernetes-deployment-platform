package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentStatsService provides statistics about Kubernetes deployments
type DeploymentStatsService struct{}

// NewDeploymentStatsService creates a new deployment stats service
func NewDeploymentStatsService() *DeploymentStatsService {
	return &DeploymentStatsService{}
}

// GetDeploymentStats returns statistics about deployments in the specified namespace
func (s *DeploymentStatsService) GetDeploymentStats(namespace string) (dto.DeploymentStatsResponse, error) {
	// Create Kubernetes client
	client, err := kubernetes.NewClient()
	if err != nil {
		return dto.DeploymentStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get deployments
	deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return dto.DeploymentStatsResponse{}, fmt.Errorf("failed to list deployments: %v", err)
	}

	// Create deployment stats array
	deploymentStats := make([]dto.DeploymentStats, 0, len(deployments.Items))
	
	for _, deployment := range deployments.Items {
		// Calculate status 
		status := "Available"
		if deployment.Status.UnavailableReplicas > 0 {
			status = "Partially Available"
		}
		if deployment.Status.AvailableReplicas == 0 {
			status = "Unavailable"
		}
		
		// Calculate rollout status
		rolloutStatus := "Complete"
		if deployment.Status.UpdatedReplicas < deployment.Status.Replicas {
			rolloutStatus = "In Progress"
		}
		
		// Add stats
		deploymentStats = append(deploymentStats, dto.DeploymentStats{
			Name:           deployment.Name,
			Namespace:      deployment.Namespace,
			Status:         status,
			RolloutStatus:  rolloutStatus,
			Replicas:       deployment.Status.Replicas,
			Updated:        deployment.Status.UpdatedReplicas,
			Ready:          deployment.Status.ReadyReplicas,
			Available:      deployment.Status.AvailableReplicas,
			Unavailable:    deployment.Status.UnavailableReplicas,
			Strategy:       string(deployment.Spec.Strategy.Type),
			Created:        deployment.CreationTimestamp.Format(time.RFC3339),
			ContainerCount: len(deployment.Spec.Template.Spec.Containers),
		})
	}
	
	return dto.DeploymentStatsResponse{
		Deployments: deploymentStats,
	}, nil
}
