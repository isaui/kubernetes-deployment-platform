package dto

// IngressRule represents a rule in a Kubernetes Ingress
type IngressRule struct {
	Host    string `json:"host"`
	Path    string `json:"path,omitempty"`
	PathType string `json:"pathType,omitempty"`
	Service  string `json:"service"`
	Port     int32  `json:"port"`
}

// IngressTLS represents TLS configuration for an Ingress
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressStats represents processed statistics for a Kubernetes ingress resource
type IngressStats struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Class       string        `json:"class,omitempty"`
	Rules       []IngressRule `json:"rules"`
	TLS         []IngressTLS  `json:"tls,omitempty"`
	Address     string        `json:"address,omitempty"`
	Created     string        `json:"created"`
	Annotations []string      `json:"annotations,omitempty"`
}

// IngressStatsResponse represents the response for ingress statistics API
type IngressStatsResponse struct {
	Ingresses []IngressStats `json:"ingresses"`
}
