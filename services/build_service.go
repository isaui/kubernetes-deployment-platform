package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/pendeploy-simple/models"
)

// BuildService handles Docker build and push operations
type BuildService struct{}

// NewBuildService creates a new BuildService instance
func NewBuildService() *BuildService {
	return &BuildService{}
}

// BuildAndPushImage builds Docker image with build args and pushes to registry
func (b *BuildService) BuildAndPushImage(cloneDir string, env map[string]string) (*models.BuildResult, error) {
	imageRegistry := env["IMAGE_REGISTRY"]
	log.Printf("🔨 Building image: %s", imageRegistry)
	
	// Build Docker image with build args
	buildResult, err := b.buildImage(cloneDir, imageRegistry, env)
	if err != nil {
		return buildResult, err
	}
	
	log.Printf("📤 Pushing image to registry: %s", imageRegistry)
	
	// Push image to registry
	pushResult, err := b.pushImage(imageRegistry)
	if err != nil {
		return &models.BuildResult{
			ImageName: imageRegistry,
			Success:   false,
			Output:    buildResult.Output + "\n" + pushResult,
			Error:     err,
		}, err
	}
	
	log.Printf("✅ Image built and pushed successfully: %s", imageRegistry)
	
	return &models.BuildResult{
		ImageName: imageRegistry,
		Success:   true,
		Output:    buildResult.Output + "\n" + pushResult,
	}, nil
}

// buildImage builds Docker image with build arguments
func (b *BuildService) buildImage(cloneDir, imageName string, env map[string]string) (*models.BuildResult, error) {
	// Prepare build command
	args := []string{"build"}
	
	// Add build args (exclude IMAGE_REGISTRY from build args)
	for key, value := range env {
		if key != "IMAGE_REGISTRY" {
			args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	// Add image tag and context
	args = append(args, "-t", imageName, ".")
	
	log.Printf("🔨 Build command: nerdctl %s", strings.Join(args, " "))
	
	// Execute build
	cmd := exec.Command("nerdctl", args...)
	cmd.Dir = cloneDir
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return &models.BuildResult{
			ImageName: imageName,
			Success:   false,
			Output:    string(output),
			Error:     fmt.Errorf("docker build failed: %w", err),
		}, err
	}
	
	return &models.BuildResult{
		ImageName: imageName,
		Success:   true,
		Output:    string(output),
	}, nil
}

// pushImage pushes the built image to registry
func (b *BuildService) pushImage(imageName string) (string, error) {
	cmd := exec.Command("nerdctl", "push", imageName)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("docker push failed: %w", err)
	}
	
	return string(output), nil
}

// VerifyImageExists checks if the image exists locally
func (b *BuildService) VerifyImageExists(imageName string) bool {
	cmd := exec.Command("nerdctl", "images", "-q", imageName)
	output, err := cmd.CombinedOutput()
	
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return false
	}
	
	return true
}