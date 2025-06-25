package dto

// ClusterVersion represents version information about the Kubernetes cluster
type ClusterVersion struct {
	GitVersion string `json:"gitVersion"`
	Platform   string `json:"platform"`
	GoVersion  string `json:"goVersion"`
	BuildDate  string `json:"buildDate"`
}

// ClusterStats represents statistics about the Kubernetes cluster
type ClusterStats struct {
	NodeCount      int `json:"nodeCount"`
	NamespaceCount int `json:"namespaceCount"`
	PodCount       int `json:"podCount"`
}

// ClusterInfoResponse represents general information about a Kubernetes cluster
type ClusterInfoResponse struct {
	Version ClusterVersion `json:"version"`
	Stats   ClusterStats   `json:"stats"`
}
