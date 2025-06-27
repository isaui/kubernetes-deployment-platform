package dto

// GitDeployRequest represents a request to deploy from a Git repository
type GitDeployRequest struct {
	ServiceID     string `json:"serviceId" binding:"required"` // ID of the service to deploy
	APIKey        string `json:"apiKey" binding:"required"`    // API Key for authentication
	CommitID      string `json:"commitId"`                     // Git commit SHA/ID to deploy (if empty, latest from default branch)
	CommitMessage string `json:"commitMessage"`                // Optional override for Git commit message to deploy
	CallbackUrl   string `json:"callbackUrl"`                 // Optional webhook URL to call on deployment success/failure
}

// GitDeployResponse represents the response for a Git deployment request
type GitDeployResponse struct {
	DeploymentID string `json:"deploymentId"`      // Generated deployment ID
	ServiceID    string `json:"serviceId"`         // Service ID from request
	Status       string `json:"status"`            // Initial status (e.g., "building")
	JobName      string `json:"jobName"`           // Name of the Kubernetes job created
	Message      string `json:"message"`           // Additional human-readable information
	CreatedAt    string `json:"createdAt"`         // Timestamp when deployment was created
}