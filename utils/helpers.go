package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateUniqueID generates a unique identifier for deployments
func GenerateUniqueID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", time.Now().Unix()+int64(rand.Intn(1000)))
}

// ExtractRepoName extracts repository name from GitHub URL
func ExtractRepoName(githubUrl string) string {
	// Remove .git suffix if present
	url := strings.TrimSuffix(githubUrl, ".git")
	
	// Split by /
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "unknown-repo"
	}
	
	// Return last part (repo name)
	return parts[len(parts)-1]
}

// CreateTempDir creates a temporary directory for build operations
func CreateTempDir(prefix string) (string, error) {
	baseDir := os.TempDir()
	uniqueDir := filepath.Join(baseDir, fmt.Sprintf("%s-%s", prefix, GenerateUniqueID()))
	
	err := os.MkdirAll(uniqueDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	return uniqueDir, nil
}

// CleanupDir removes a directory and all its contents
func CleanupDir(dir string) error {
	if dir == "" || dir == "/" || dir == os.TempDir() {
		return fmt.Errorf("refusing to delete important directory: %s", dir)
	}
	
	return os.RemoveAll(dir)
}

// ValidateImageRegistry validates that IMAGE_REGISTRY is present in env
func ValidateImageRegistry(env map[string]string) error {
	imageRegistry, exists := env["IMAGE_REGISTRY"]
	if !exists {
		return fmt.Errorf("IMAGE_REGISTRY is required in env")
	}
	
	if strings.TrimSpace(imageRegistry) == "" {
		return fmt.Errorf("IMAGE_REGISTRY cannot be empty")
	}
	
	return nil
}

// SanitizeAppName creates a valid Kubernetes app name from repo name
func SanitizeAppName(repoName string) string {
	// Convert to lowercase and replace invalid characters
	name := strings.ToLower(repoName)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	
	// Remove invalid characters
	var result strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	
	finalName := result.String()
	
	// Ensure it doesn't start or end with hyphen
	finalName = strings.Trim(finalName, "-")
	
	// Ensure it's not empty
	if finalName == "" {
		finalName = "app-" + GenerateUniqueID()
	}
	
	return finalName
}