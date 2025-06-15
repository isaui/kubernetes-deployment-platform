package models

// DeploymentRequest represents the incoming deployment request
type DeploymentRequest struct {
	GithubUrl string            `json:"githubUrl" binding:"required"`
	Env       map[string]string `json:"env" binding:"required"`
}

// DeploymentResponse represents the deployment response
type DeploymentResponse struct {
	Status     string `json:"status"`
	ImageName  string `json:"imageName"`
	ServiceURL string `json:"serviceUrl,omitempty"`
	Message    string `json:"message"`
	BuildLogs  string `json:"buildLogs,omitempty"`
}

// BuildResult represents the result of a build operation
type BuildResult struct {
	ImageName string
	Success   bool
	Output    string
	Error     error
}

// GitResult represents the result of a git operation
type GitResult struct {
	CloneDir string
	Success  bool
	Output   string
	Error    error
}

// KubernetesResult represents the result of kubernetes operations
type KubernetesResult struct {
	Success bool
	Output  string
	Error   error
}