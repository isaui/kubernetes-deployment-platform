package dto

// NodeResource represents a Kubernetes node resource (CPU, Memory, Storage)
type NodeResource struct {
	Capacity    string  `json:"capacity"`
	Allocatable string  `json:"allocatable"`
	Usage       string  `json:"usage,omitempty"`       // Actual usage value (e.g., "380m" for CPU, "12421Mi" for memory)
	Percentage  float64 `json:"percentage,omitempty"` // Usage as percentage of capacity
}

// NodeCondition represents a Kubernetes node condition
type NodeCondition struct {
	Status            string `json:"status"`
	Reason            string `json:"reason,omitempty"`
	Message           string `json:"message,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
}

// NodeConditions is a map of condition types to their details
type NodeConditions map[string]NodeCondition

// NodeStats represents processed statistics for a Kubernetes node
type NodeStats struct {
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	Conditions     NodeConditions `json:"conditions,omitempty"`
	Roles          []string       `json:"roles"`
	Created        string         `json:"created"`
	KubeletVersion string         `json:"kubeletVersion,omitempty"`
	OSImage        string         `json:"osImage,omitempty"`
	CPU            NodeResource   `json:"cpu"`
	Memory         NodeResource   `json:"memory"`
	Storage        NodeResource   `json:"storage,omitempty"`
}

// NodeStatsResponse represents the response for node statistics API
type NodeStatsResponse struct {
	Nodes []NodeStats `json:"nodes"`
}
