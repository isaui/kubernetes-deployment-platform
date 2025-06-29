package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
)

// NewRegistryAPI creates a new registry API client
func NewRegistryAPI(baseURL string) *dto.RegistryAPI {
	// Extract service info from URL for compatibility
	serviceName, namespace := parseServiceFromURL(baseURL)
	
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to create Kubernetes client: %v", err)
	}

	return &dto.RegistryAPI{
		ServiceName: serviceName,
		Namespace:   namespace,
		K8sClient:   k8sClient,
	}
}

// NewRegistryAPIFromRegistry creates a registry API client from registry URL
func NewRegistryAPIFromRegistry(registryURL string) (*dto.RegistryAPI, error) {
	serviceName, namespace := parseServiceFromRegistryURL(registryURL)
	
	log.Printf("Creating Registry API client for service: %s in namespace: %s", serviceName, namespace)
	
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	return &dto.RegistryAPI{
		ServiceName: serviceName,
		Namespace:   namespace,
		K8sClient:   k8sClient,
	}, nil
}

// parseServiceFromRegistryURL extracts service name and namespace from registry URL
func parseServiceFromRegistryURL(registryURL string) (string, string) {
	// Clean URL from protocol
	cleaned := strings.TrimPrefix(registryURL, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")
	
	// Extract service name from URL like: registry-abc123.registry.svc.cluster.local:5000
	parts := strings.Split(cleaned, ".")
	if len(parts) > 0 {
		serviceName := parts[0]
		// Remove port if present
		if idx := strings.LastIndex(serviceName, ":"); idx > 0 {
			serviceName = serviceName[:idx]
		}
		return serviceName, "registry" // Default namespace
	}
	
	return "registry", "registry"
}

// parseServiceFromURL extracts service info from any URL format
func parseServiceFromURL(url string) (string, string) {
	// Default values
	return "registry", "registry"
}

// proxyRequest makes an HTTP request via Kubernetes service proxy
func proxyRequest(ctx context.Context, method, path string, api *dto.RegistryAPI) (*http.Response, error) {
	if api.K8sClient == nil {
		return nil, fmt.Errorf("kubernetes client not available")
	}

	log.Printf("Making proxy request: %s %s via service %s.%s", method, path, api.ServiceName, api.Namespace)

	// Use Kubernetes client service proxy - simplified without scheme/port complexity
	req := api.K8sClient.Clientset.CoreV1().Services(api.Namespace).ProxyGet(
		"http", // No scheme - let K8s handle it
		api.ServiceName,
		"5000", // No port - let K8s handle it  
		path,
		map[string]string{},
	)

	// Execute request and get raw response
	body, err := req.DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %v", err)
	}

	// Create mock HTTP response for compatibility
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     make(http.Header),
	}

	return resp, nil
}

// Response structs
type catalogResponse struct {
	Repositories []string `json:"repositories"`
}



// TestConnection tests if the registry is accessible
func TestConnection(ctx context.Context, api *dto.RegistryAPI) error {
	log.Printf("Testing registry connection to service: %s in namespace: %s", api.ServiceName, api.Namespace)
	
	resp, err := proxyRequest(ctx, "GET", "v2/", api)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}
	
	log.Printf("Registry connection test successful")
	return nil
}

// GetRepositories returns all repositories in the registry
func GetRepositories(ctx context.Context, api *dto.RegistryAPI) ([]string, error) {
	log.Printf("Getting repositories from service: %s", api.ServiceName)
	
	resp, err := proxyRequest(ctx, "GET", "v2/_catalog", api)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %v", err)
	}
	defer resp.Body.Close()
	
	log.Printf("Registry catalog response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get catalog: status %d", resp.StatusCode)
	}
	
	var catalog catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, fmt.Errorf("failed to decode catalog: %v", err)
	}
	
	log.Printf("Successfully retrieved %d repositories", len(catalog.Repositories))
	return catalog.Repositories, nil
}

// GetTags returns all tags for a repository
func GetTags(ctx context.Context, api *dto.RegistryAPI, repository string) ([]string, error) {
	path := fmt.Sprintf("v2/%s/tags/list", repository)
	
	resp, err := proxyRequest(ctx, "GET", path, api)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tags: status %d", resp.StatusCode)
	}
	
	var tags dto.TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %v", err)
	}
	
	return tags.Tags, nil
}

// GetManifest returns the manifest for a specific image tag
func GetManifest(ctx context.Context, api *dto.RegistryAPI, repository, tag string) (*dto.ManifestResponse, error) {
	path := fmt.Sprintf("v2/%s/manifests/%s", repository, tag)
	
	resp, err := proxyRequest(ctx, "GET", path, api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manifest: status %d", resp.StatusCode)
	}
	
	var manifest dto.ManifestResponse
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	
	return &manifest, nil
}

// GetImageInfo returns details about an image
func GetImageInfo(ctx context.Context, api *dto.RegistryAPI, repository string, tags []string) (dto.RegistryImageInfo, error) {
	if len(tags) == 0 {
		return dto.RegistryImageInfo{
			Name:      repository,
			Tags:      []string{},
			Size:      0,
			CreatedAt: time.Now(),
		}, nil
	}
	
	// Get manifest for the first tag to determine size
	manifest, err := GetManifest(ctx, api, repository, tags[0])
	if err != nil {
		// Return partial info if we can't get manifest
		return dto.RegistryImageInfo{
			Name:      repository,
			Tags:      tags,
			Size:      0,
			CreatedAt: time.Now(),
		}, nil
	}
	
	// Calculate total size of all layers
	var totalSize int64
	for _, layer := range manifest.Layers {
		totalSize += layer.Size
	}
	
	// Add config size
	totalSize += manifest.Config.Size
	
	return dto.RegistryImageInfo{
		Name:      repository,
		Tags:      tags,
		Size:      totalSize,
		CreatedAt: time.Now(),
	}, nil
}

// ListImages returns a list of all images in the registry with their details
func ListImages(ctx context.Context, api *dto.RegistryAPI) ([]dto.RegistryImageInfo, error) {
	// Get list of repositories
	repositories, err := GetRepositories(ctx, api)
	if err != nil {
		return nil, err
	}
	
	var images []dto.RegistryImageInfo
	
	// Get details for each repository
	for _, repo := range repositories {
		// Get tags for this repository
		tags, err := GetTags(ctx, api, repo)
		if err != nil {
			log.Printf("Failed to get tags for repository %s: %v", repo, err)
			// Add repository with empty tags if we can't get tags
			images = append(images, dto.RegistryImageInfo{
				Name:      repo,
				Tags:      []string{},
				Size:      0,
				CreatedAt: time.Now(),
			})
			continue
		}
		
		// Get image info
		imageInfo, err := GetImageInfo(ctx, api, repo, tags)
		if err != nil {
			log.Printf("Failed to get full info for repository %s: %v", repo, err)
			images = append(images, dto.RegistryImageInfo{
				Name:      repo,
				Tags:      tags,
				Size:      0,
				CreatedAt: time.Now(),
			})
			continue
		}
		
		images = append(images, imageInfo)
	}
	
	return images, nil
}

// ParseRegistryURL ensures the URL is properly formatted for API calls (simplified for internal use)
func ParseRegistryURL(registryURL string) (string, error) {
	// Since we're using K8s proxy, just return the URL as-is for compatibility
	// The actual service parsing is done in the constructor functions
	log.Printf("Parsing registry URL for K8s proxy: %s", registryURL)
	return registryURL, nil
}