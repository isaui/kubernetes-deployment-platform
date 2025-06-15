package models

// DeploymentRequest struktur untuk permintaan deployment
type DeploymentRequest struct {
	CommitID  string `json:"commitId" binding:"required"`
	Branch    string `json:"branch" binding:"required"`
	RepoURL   string `json:"repoUrl" binding:"required"`
}
