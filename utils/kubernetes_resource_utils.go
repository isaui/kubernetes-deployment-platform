package utils

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	k8s "github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetResourceName generates a consistent, immutable resource name based on service ID
// This ensures resources can be tracked even if service name changes
func GetResourceName(service models.Service) string {
	// Use service ID (UUID) as the resource name, but ensure it's DNS-1035 compliant
	// UUID needs to be prefixed with a letter to be valid in Kubernetes
	// Add 's-' prefix to ensure it starts with a letter (DNS-1035 compliance)
	return "s-" + service.ID
}
func getMainContainerName() string {
	return "app"
}
// SanitizeLabel makes a string valid for use as a Kubernetes label value
// by replacing invalid characters with '-' and ensuring it meets label requirements
func SanitizeLabel(value string) string {
	// Convert to lowercase first
	value = strings.ToLower(value)
	
	// Replace spaces and other invalid characters with dashes
	reg := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)
	sanitized := reg.ReplaceAllString(value, "-")
	
	// Ensure it starts with an alphanumeric character
	if len(sanitized) > 0 && !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(sanitized) {
		sanitized = "x" + sanitized
	}
	
	// Ensure it ends with an alphanumeric character
	if len(sanitized) > 0 && !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(sanitized) {
		sanitized = sanitized + "x"
	}
	
	// Truncate if it's too long (63 is the Kubernetes limit for label values)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
		// Ensure it still ends with an alphanumeric character after truncation
		if !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(sanitized) {
			sanitized = sanitized[:62] + "x"
		}
	}
	
	return sanitized
}

// GetDefaultDomainName extracts repository name from git URL to create a default domain name
func GetDefaultDomainName(service models.Service) string {
	// Extract repo name from Git URL
	repoName := extractRepoNameFromURL(service.RepoURL)
	
	// Create sanitized parts for the domain name
	sanitizedRepoName := SanitizeLabel(repoName)
	sanitizedBranch := SanitizeLabel(service.Branch)
	
	// Default to 'main' if branch is empty
	if sanitizedBranch == "" {
		sanitizedBranch = "main"
	}
	
	// Truncate environmentID to 6 characters to keep hostnames shorter
	shortEnvID := service.EnvironmentID
	if len(shortEnvID) > 6 {
		shortEnvID = shortEnvID[:6]
	}
	
	// Format: repo-name-branch.env-id.app.isacitra.com
	return fmt.Sprintf("%s-%s.%s.app.isacitra.com", 
		sanitizedRepoName, 
		sanitizedBranch, 
		shortEnvID)
}

// extractRepoNameFromURL extracts the repository name from a git URL
func extractRepoNameFromURL(repoURL string) string {
	// Handle empty URL case
	if repoURL == "" {
		return "app"
	}
	
	// Remove trailing .git if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Extract repo name from different URL formats
	// Case 1: https://github.com/username/repo
	if strings.Contains(repoURL, "github.com") || strings.Contains(repoURL, "gitlab.com") {
		parts := strings.Split(repoURL, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1] // Get last part after slash
		}
	}
	
	// Case 2: git@github.com:username/repo
	if strings.Contains(repoURL, "@") && strings.Contains(repoURL, ":") {
		parts := strings.Split(repoURL, ":")
		if len(parts) > 1 {
			repoPath := parts[len(parts)-1]
			repoPathParts := strings.Split(repoPath, "/")
			return repoPathParts[len(repoPathParts)-1] // Get last part
		}
	}
	
	// Default case: just use the whole URL as a basis and sanitize it
	parts := strings.Split(repoURL, "/")
	return parts[len(parts)-1]
}

// GetResourceLabels generates consistent labels for resources
func GetResourceLabels(service models.Service) map[string]string {
	return map[string]string{
		"app":         GetResourceName(service), // Use immutable resource name
		"service-id":  service.ID,
		"service-name": SanitizeLabel(service.Name), // Sanitize name for Kubernetes label compliance
		"environment": service.EnvironmentID,
		"managed-by":  "pendeploy",
	}
}
// GetKubernetesResourceStatus gets the status of all resources for a service via Kubernetes API
func GetKubernetesResourceStatus(service models.Service) (map[string]interface{}, error) {
	// Create Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
    log.Println("Kubernetes client created successfully")
	ctx := context.Background()
	resourceName := GetResourceName(service) // This is just service.ID
	status := make(map[string]interface{})

	// Get Deployment status
	deployment, err := k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["deployment"] = map[string]interface{}{
			"name":              deployment.Name,
			"replicas":          deployment.Status.Replicas,
			"readyReplicas":     deployment.Status.ReadyReplicas,
			"availableReplicas": deployment.Status.AvailableReplicas,
			"conditions":        deployment.Status.Conditions,
		}
	}
    log.Println("Deployment status retrieved successfully")
	// Get Service status
	svc, err := k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["service"] = map[string]interface{}{
			"name":      svc.Name,
			"clusterIP": svc.Spec.ClusterIP,
			"ports":     svc.Spec.Ports,
		}
	}

	// Get Ingress status
	ingress, err := k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["ingress"] = map[string]interface{}{
			"name":  ingress.Name,
			"hosts": ingress.Spec.Rules,
		}
	}	
    log.Println("Ingress status retrieved successfully")
	// Get HPA status if exists
	hpa, err := k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Get(ctx, resourceName, metav1.GetOptions{})
	if err == nil {
		status["hpa"] = map[string]interface{}{
			"name":           hpa.Name,
			"currentReplicas": hpa.Status.CurrentReplicas,
			"desiredReplicas": hpa.Status.DesiredReplicas,
		}
	}
    log.Println("HPA status retrieved successfully")
	return status, nil
}


