package models

// BuildImageResult berisi informasi hasil build image
type BuildImageResult struct {
	ImageName   string `json:"imageName"`
	ImageTag    string `json:"imageTag"`
	ContainerID string `json:"containerID"` 
	Output      string `json:"output"`
}

// ContainerRunResult berisi informasi hasil running container
type ContainerRunResult struct {
	ContainerID string `json:"containerId"`
	ImageName   string `json:"imageName"`
	ImageTag    string `json:"imageTag"`
	Port        int    `json:"port"`
	Output      string `json:"output"`
}
