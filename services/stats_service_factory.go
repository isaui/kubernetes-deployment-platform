package services

import (
	"github.com/pendeploy-simple/dto"
)

// StatsService is a lightweight façade that coordinates stats operations
// Each method creates its own specialized service instance when needed
type StatsService struct {}

// NewStatsService creates a new stats service
func NewStatsService() (*StatsService, error) {
	return &StatsService{}, nil
}

// GetPodStats returns statistics about pods, creating a new PodStatsService instance
func (s *StatsService) GetPodStats(namespace string) (dto.PodStatsResponse, error) {
	podService := NewPodStatsService()
	return podService.GetPodStats(namespace)
}

// GetNodeStats returns statistics about nodes, creating a new NodeStatsService instance
func (s *StatsService) GetNodeStats() (dto.NodeStatsResponse, error) {
	nodeService := NewNodeStatsService()
	return nodeService.GetNodeStats()
}

// GetDeploymentStats returns stats about deployments in the cluster
func (s *StatsService) GetDeploymentStats(namespace string) (dto.DeploymentStatsResponse, error) {
	deploymentService := NewDeploymentStatsService()
	return deploymentService.GetDeploymentStats(namespace)
}

// GetServiceStats returns statistics about Kubernetes services
func (s *StatsService) GetServiceStats(namespace string) (dto.ServiceStatsResponse, error) {
	serviceStatsService := NewServiceStatsService()
	return serviceStatsService.GetServiceStats(namespace)
}

// GetIngressStats returns statistics about Kubernetes ingress resources
func (s *StatsService) GetIngressStats(namespace string) (dto.IngressStatsResponse, error) {
	ingressStatsService := NewIngressStatsService()
	return ingressStatsService.GetIngressStats(namespace)
}

// GetCertificateStats returns statistics about cert-manager certificates
func (s *StatsService) GetCertificateStats(namespace string) (dto.CertificateStatsResponse, error) {
	certificateStatsService := NewCertificateStatsService()
	return certificateStatsService.GetCertificateStats(namespace)
}

// GetClusterInfo returns general information about the cluster
func (s *StatsService) GetClusterInfo() (dto.ClusterInfoResponse, error) {
	clusterInfoService := NewClusterInfoService()
	return clusterInfoService.GetClusterInfo()
}
