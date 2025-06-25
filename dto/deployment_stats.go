package dto

import ()

// DeploymentStats represents processed statistics for a Kubernetes deployment
type DeploymentStats struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Status         string `json:"status"`
	RolloutStatus  string `json:"rolloutStatus"`
	Replicas       int32  `json:"replicas"`
	Updated        int32  `json:"updated"`
	Ready          int32  `json:"ready"`
	Available      int32  `json:"available"`
	Unavailable    int32  `json:"unavailable"`
	Strategy       string `json:"strategy"`
	Created        string `json:"created"`
	ContainerCount int    `json:"containerCount"`
}

// DeploymentStatsResponse represents the response for deployment statistics API
type DeploymentStatsResponse struct {
	Deployments []DeploymentStats `json:"deployments"`
}
