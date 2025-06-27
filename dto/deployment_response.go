package dto

import (
	"time"

	"github.com/pendeploy-simple/models"
)

// DeploymentResponse represents a deployment response
type DeploymentResponse struct {
	ID            string    `json:"id"`
	ServiceID     string    `json:"serviceId"`
	Status        string    `json:"status"`
	CommitSHA     string    `json:"commitSha"`
	CommitMessage string    `json:"commitMessage"`
	Image         string    `json:"image"`
	Version       string    `json:"version"`
	CreatedAt     time.Time `json:"createdAt"`
}

// NewDeploymentResponseFromModel creates a new DeploymentResponse from a models.Deployment
func NewDeploymentResponseFromModel(deployment models.Deployment) DeploymentResponse {
	return DeploymentResponse{
		ID:            deployment.ID,
		ServiceID:     deployment.ServiceID,
		Status:        string(deployment.Status),
		CommitSHA:     deployment.CommitSHA,
		CommitMessage: deployment.CommitMessage,
		Image:         deployment.Image,
		Version:       deployment.Version,
		CreatedAt:     deployment.CreatedAt,
	}
}

// ResourceStatusResponse represents the status of Kubernetes resources for a service
type ResourceStatusResponse struct {
	Deployment *DeploymentStatusInfo `json:"deployment,omitempty"`
	Service    *ServiceStatusInfo    `json:"service,omitempty"`
	Ingress    *IngressStatusInfo    `json:"ingress,omitempty"`
	HPA        *HPAStatusInfo        `json:"hpa,omitempty"`
}

// DeploymentStatusInfo contains status information for a Kubernetes Deployment
type DeploymentStatusInfo struct {
	Name              string `json:"name"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Replicas          int32  `json:"replicas"`
	Age               string `json:"age"`
	Image             string `json:"image"`
}

// ServiceStatusInfo contains status information for a Kubernetes Service
type ServiceStatusInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIp"`
	Ports     string `json:"ports"`
	Age       string `json:"age"`
}

// IngressStatusInfo contains status information for a Kubernetes Ingress
type IngressStatusInfo struct {
	Name   string   `json:"name"`
	Hosts  []string `json:"hosts"`
	TLS    bool     `json:"tls"`
	Age    string   `json:"age"`
	Status string   `json:"status"`
}

// HPAStatusInfo contains status information for a Kubernetes HorizontalPodAutoscaler
type HPAStatusInfo struct {
	Name           string `json:"name"`
	MinReplicas    int32  `json:"minReplicas"`
	MaxReplicas    int32  `json:"maxReplicas"`
	CurrentReplicas int32 `json:"currentReplicas"`
	TargetCPU      int32  `json:"targetCpu"`
	CurrentCPU     int32  `json:"currentCpu"`
	Age            string `json:"age"`
}
