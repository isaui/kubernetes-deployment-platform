package dto

// PodResource represents resource stats for a pod (CPU, Memory)
type PodResource struct {
	Usage      string  `json:"usage"`       // Actual usage value (e.g., "0.15" for CPU, "120Mi" for memory)
	Request    string  `json:"request"`     // Resource request (min)
	Limit      string  `json:"limit"`       // Resource limit (max)
	Percentage float64 `json:"percentage"`  // Usage as percentage of limit (or request if limit not set)
}

// PodStats represents processed statistics for a Kubernetes pod
type PodStats struct {
	Name           string      `json:"name"`
	Namespace      string      `json:"namespace"`
	Status         string      `json:"status"`
	Type           string      `json:"type"`           // Deployment, StatefulSet, etc
	ContainerCount int         `json:"containerCount"`
	CPU            PodResource `json:"cpu"`
	Memory         PodResource `json:"memory"`
	Created        string      `json:"created"`
}

// PodStatsResponse represents the response for pod statistics API
type PodStatsResponse struct {
	Pods []PodStats `json:"pods"`
}
