package services

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/utils"
)

// GitService handles git operations
type GitService struct{}

// NewGitService creates a new GitService instance
func NewGitService() *GitService {
	return &GitService{}
}

// CloneRepository clones a GitHub repository to a temporary directory
func (g *GitService) CloneRepository(githubUrl string) (*models.GitResult, error) {
	log.Printf("🔄 Cloning repository: %s", githubUrl)
	
	// Create temporary directory
	repoName := utils.ExtractRepoName(githubUrl)
	cloneDir, err := utils.CreateTempDir(fmt.Sprintf("deploy-%s", repoName))
	if err != nil {
		return &models.GitResult{
			Success: false,
			Error:   fmt.Errorf("failed to create temp directory: %w", err),
		}, err
	}
	
	// Execute git clone
	cmd := exec.Command("git", "clone", "--depth", "1", githubUrl, cloneDir)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		// Cleanup on failure
		utils.CleanupDir(cloneDir)
		return &models.GitResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Errorf("git clone failed: %w", err),
		}, err
	}
	
	// Verify Dockerfile exists
	dockerfilePath := filepath.Join(cloneDir, "Dockerfile")
	if !g.fileExists(dockerfilePath) {
		utils.CleanupDir(cloneDir)
		return &models.GitResult{
			Success: false,
			Error:   fmt.Errorf("Dockerfile not found in repository root"),
		}, fmt.Errorf("Dockerfile not found")
	}
	
	// Verify kubernetes directory exists
	kubernetesDir := filepath.Join(cloneDir, "kubernetes")
	if !g.dirExists(kubernetesDir) {
		utils.CleanupDir(cloneDir)
		return &models.GitResult{
			Success: false,
			Error:   fmt.Errorf("kubernetes/ directory not found in repository"),
		}, fmt.Errorf("kubernetes directory not found")
	}
	
	log.Printf("✅ Repository cloned successfully to: %s", cloneDir)
	
	return &models.GitResult{
		CloneDir: cloneDir,
		Success:  true,
		Output:   string(output),
	}, nil
}

// fileExists checks if a file exists
func (g *GitService) fileExists(path string) bool {
	if _, err := exec.Command("test", "-f", path).Output(); err != nil {
		return false
	}
	return true
}

// dirExists checks if a directory exists
func (g *GitService) dirExists(path string) bool {
	if _, err := exec.Command("test", "-d", path).Output(); err != nil {
		return false
	}
	return true
}