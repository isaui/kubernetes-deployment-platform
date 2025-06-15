package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isabu/pendeploy-handal/models"
	"github.com/isabu/pendeploy-handal/utils"
)

// DeploymentService menangani operasi terkait deployment
type DeploymentService struct {
	// Nantinya bisa ditambahkan dependensi ke repository, logger, dll.
	BaseDir string
}

// NewDeploymentService membuat instance baru dari DeploymentService
func NewDeploymentService() *DeploymentService {
	// Buat base directory jika belum ada
	baseDir := filepath.Join(os.TempDir(), "pendeploy-handal")
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		log.Printf("Failed to create base directory: %v", err)
	}
	
	return &DeploymentService{
		BaseDir: baseDir,
	}
}

// CloneRepository melakukan git clone dari repositori dengan branch tertentu
func (s *DeploymentService) CloneRepository(repoURL, branch string) (string, error) {
	// Dapatkan direktori repository
	repoDir := s.GetRepoDirectory(repoURL)
	
	// Hapus direktori jika sudah ada (selalu fresh clone)
	if _, err := os.Stat(repoDir); err == nil {
		log.Printf("Repository directory exists, removing for fresh clone: %s", repoDir)
		if err := os.RemoveAll(repoDir); err != nil {
			log.Printf("Failed to remove existing directory: %v", err)
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}
	
	// Clone repository dengan branch yang ditentukan
	log.Printf("Cloning repository %s branch %s to %s", repoURL, branch, repoDir)
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, repoDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}
	
	return fmt.Sprintf("Repository cloned successfully to %s\n%s", repoDir, string(output)), nil
}

// GetRepoDirectory mengembalikan direktori repo dari URL
func (s *DeploymentService) GetRepoDirectory(repoURL string) string {
	repoName := utils.ExtractRepoName(repoURL)
	return filepath.Join(s.BaseDir, utils.SanitizeDirName(repoName))
}

// BuildImageResult dan ContainerRunResult sudah dipindahkan ke models/deployment_result.go

// BuildImage membuat container image menggunakan nerdctl dari Dockerfile di repository
func (s *DeploymentService) BuildImage(repoURL, commitID, branch string) (*models.BuildImageResult, error) {
	// Dapatkan direktori repository
	repoDir := s.GetRepoDirectory(repoURL)
	
	// Periksa apakah Dockerfile ada
	dockerfilePath := filepath.Join(repoDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Dockerfile not found in repository root: %s", repoDir)
	}
	
	// Buat nama image yang unik dan trackable
	repoName := utils.ExtractRepoName(repoURL)
	// Gunakan full commit ID
	// Tidak perlu memotong commit ID, gunakan sepenuhnya
	
	// Buat nama image dengan format: organization/user::repository::branch:commitid
	// Karakter '::' tidak dapat digunakan dalam username/organization GitHub
	// Memisahkan parts dari URL
	repoNameParts := strings.Split(repoName, "/")
	orgUser := repoNameParts[0]
	repo := ""
	if len(repoNameParts) > 1 {
		repo = repoNameParts[1]
	} else {
		repo = "repo"
	}
	
	// Format: org-user-repo-branch:commit
	// Using dashes as separators instead of double colons, which are invalid in image names
	imageName := fmt.Sprintf("%s-%s-%s", orgUser, repo, branch)
	imageTag := commitID
	imageFullName := fmt.Sprintf("%s:%s", imageName, imageTag)
	
	// Jalankan nerdctl build dengan Dockerfile
	cmd := exec.Command("nerdctl", "build", "-t", imageFullName, "-f", "Dockerfile", ".")
	// Set working directory ke repo
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nerdctl build failed: %w\n%s", err, string(output))
	}
	
	// Generate a container ID berdasarkan format name
	// Format: org-repo-branch-commit
	containerID := fmt.Sprintf("%s-%s-%s-%s", orgUser, repo, branch, commitID)
	
	// Kembalikan hasil build
	return &models.BuildImageResult{
		ImageName:   imageName,
		ImageTag:    imageTag,
		ContainerID: containerID,
		Output:      string(output),
	}, nil
}

// Kita tidak memerlukan RunContainer karena deployment akan dilakukan via Kubernetes
// Fungsi ini dihapus sesuai dengan kebutuhan fokus pada Kubernetes
