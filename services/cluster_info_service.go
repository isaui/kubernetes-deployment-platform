package services

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterInfoService provides general information about Kubernetes cluster
type ClusterInfoService struct{}

// NewClusterInfoService creates a new cluster info service
func NewClusterInfoService() *ClusterInfoService {
	return &ClusterInfoService{}
}

// GetClusterInfo returns general information about the cluster
func (s *ClusterInfoService) GetClusterInfo() (dto.ClusterInfoResponse, error) {
	// Create Kubernetes client
	client, err := kubernetes.NewClient()
	if err != nil {
		return dto.ClusterInfoResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get version information
	version, err := client.Clientset.Discovery().ServerVersion()
	if err != nil {
		return dto.ClusterInfoResponse{}, fmt.Errorf("failed to get server version: %v", err)
	}

	// Get nodes count
	nodes, err := client.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	nodeCount := 0
	if err == nil {
		nodeCount = len(nodes.Items)
	} else {
		fmt.Printf("Warning: Failed to get node count: %v\n", err)
	}

	// Get namespaces count
	namespaces, err := client.Clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	namespaceCount := 0
	if err == nil {
		namespaceCount = len(namespaces.Items)
	} else {
		fmt.Printf("Warning: Failed to get namespace count: %v\n", err)
	}

	// Get pods count across all namespaces
	pods, err := client.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	podCount := 0
	if err == nil {
		podCount = len(pods.Items)
	} else {
		fmt.Printf("Warning: Failed to get pod count: %v\n", err)
	}

	return dto.ClusterInfoResponse{
		Version: dto.ClusterVersion{
			GitVersion: version.GitVersion,
			Platform:   version.Platform,
			GoVersion:  version.GoVersion,
			BuildDate:  version.BuildDate,
		},
		Stats: dto.ClusterStats{
			NodeCount:      nodeCount,
			NamespaceCount: namespaceCount,
			PodCount:       podCount,
		},
	}, nil
}
