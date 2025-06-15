package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/utils"
)

// KubernetesService handles Kubernetes operations
type KubernetesService struct{}

// NewKubernetesService creates a new KubernetesService instance
func NewKubernetesService() *KubernetesService {
	return &KubernetesService{}
}

// ProcessAndApplyManifests processes Kubernetes manifests with environment substitution and applies them
func (k *KubernetesService) ProcessAndApplyManifests(cloneDir string, env map[string]string) (*models.KubernetesResult, error) {
	kubernetesDir := filepath.Join(cloneDir, "kubernetes")
	
	log.Printf("🎯 Processing Kubernetes manifests from: %s", kubernetesDir)
	
	// Create temporary directory for processed manifests
	processedDir, err := utils.CreateTempDir("k8s-processed")
	if err != nil {
		return &models.KubernetesResult{
			Success: false,
			Error:   fmt.Errorf("failed to create processed manifests directory: %w", err),
		}, err
	}
	defer utils.CleanupDir(processedDir)
	
	// Process manifests with environment substitution
	err = k.processManifests(kubernetesDir, processedDir, env)
	if err != nil {
		return &models.KubernetesResult{
			Success: false,
			Error:   fmt.Errorf("failed to process manifests: %w", err),
		}, err
	}
	
	// Apply processed manifests to cluster
	output, err := k.applyManifests(processedDir)
	if err != nil {
		return &models.KubernetesResult{
			Success: false,
			Output:  output,
			Error:   fmt.Errorf("failed to apply manifests: %w", err),
		}, err
	}
	
	log.Printf("✅ Kubernetes manifests applied successfully")
	
	return &models.KubernetesResult{
		Success: true,
		Output:  output,
	}, nil
}

// processManifests processes YAML files with environment variable substitution
func (k *KubernetesService) processManifests(sourceDir, targetDir string, env map[string]string) error {
	// Find all YAML files
	yamlFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find YAML files: %w", err)
	}
	
	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.yml"))
	if err == nil {
		yamlFiles = append(yamlFiles, ymlFiles...)
	}
	
	if len(yamlFiles) == 0 {
		return fmt.Errorf("no YAML files found in kubernetes directory")
	}
	
	log.Printf("📝 Found %d manifest files to process", len(yamlFiles))
	
	// Process each YAML file
	for _, yamlFile := range yamlFiles {
		err := k.processYAMLFile(yamlFile, targetDir, env)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", yamlFile, err)
		}
	}
	
	return nil
}

// processYAMLFile processes a single YAML file with environment substitution
func (k *KubernetesService) processYAMLFile(sourceFile, targetDir string, env map[string]string) error {
	// Read source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Perform environment variable substitution
	processedContent := string(content)
	for key, value := range env {
		placeholder := fmt.Sprintf("${%s}", key)
		processedContent = strings.ReplaceAll(processedContent, placeholder, value)
	}
	
	// Write processed file to target directory
	fileName := filepath.Base(sourceFile)
	targetFile := filepath.Join(targetDir, fileName)
	
	err = os.WriteFile(targetFile, []byte(processedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write processed file: %w", err)
	}
	
	log.Printf("📄 Processed: %s -> %s", fileName, targetFile)
	return nil
}

// applyManifests applies all processed manifests to the Kubernetes cluster
func (k *KubernetesService) applyManifests(manifestsDir string) (string, error) {
	log.Printf("🚀 Applying manifests to Kubernetes cluster...")
	
	// Execute kubectl apply
	cmd := exec.Command("kubectl", "apply", "-f", manifestsDir)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("kubectl apply failed: %w", err)
	}
	
	return string(output), nil
}

// GetServiceURL attempts to get the service URL for the deployed application
func (k *KubernetesService) GetServiceURL(appName string) (string, error) {
	// Try to get service info
	cmd := exec.Command("kubectl", "get", "service", appName, "-o", "jsonpath={.spec.clusterIP}:{.spec.ports[0].port}")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		// Service might not exist yet or have different name
		return "", fmt.Errorf("could not get service URL: %w", err)
	}
	
	serviceInfo := strings.TrimSpace(string(output))
	if serviceInfo == "" || serviceInfo == ":" {
		return "", fmt.Errorf("service not found or not ready")
	}
	
	return fmt.Sprintf("http://%s", serviceInfo), nil
}

// VerifyDeployment checks if the deployment is ready
func (k *KubernetesService) VerifyDeployment(appName string) (bool, error) {
	cmd := exec.Command("kubectl", "get", "deployment", appName, "-o", "jsonpath={.status.readyReplicas}")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return false, err
	}
	
	readyReplicas := strings.TrimSpace(string(output))
	return readyReplicas != "" && readyReplicas != "0", nil
}