package dto

// ServicePort represents a Kubernetes service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// ServiceStats represents processed statistics for a Kubernetes service
type ServiceStats struct {
	Name          string        `json:"name"`
	Namespace     string        `json:"namespace"`
	Type          string        `json:"type"`
	ClusterIP     string        `json:"clusterIP"`
	ExternalIPs   []string      `json:"externalIPs,omitempty"`
	LoadBalancer  string        `json:"loadBalancer,omitempty"`
	Ports         []ServicePort `json:"ports"`
	Selector      []string      `json:"selector"`
	PodCount      int           `json:"podCount"`
	Created       string        `json:"created"`
	EndpointCount int           `json:"endpointCount"`
}

// ServiceStatsResponse represents the response for service statistics API
type ServiceStatsResponse struct {
	Services []ServiceStats `json:"services"`
}
