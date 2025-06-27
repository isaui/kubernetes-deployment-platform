package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pendeploy-simple/dto"
)

// RegistryAPI handles interactions with Docker Registry HTTP API v2
type RegistryAPI struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// NewRegistryAPI creates a new registry API client
func NewRegistryAPI(baseURL, username, password string) *RegistryAPI {
	// Ensure baseURL ends with /v2/
	if !strings.HasSuffix(baseURL, "/v2/") {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		baseURL += "v2/"
	}

	return &RegistryAPI{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// catalogResponse represents the response from the catalog endpoint
type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

// tagsResponse represents the response from the tags endpoint
type tagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// manifestResponse represents the response from the manifest endpoint
type manifestResponse struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        manifestConfig  `json:"config"`
	Layers        []manifestLayer `json:"layers"`
	History       []manifestItem  `json:"history,omitempty"`
}

type manifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type manifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type manifestItem struct {
	V1Compatibility string `json:"v1Compatibility"`
}

// GetRepositories returns all repositories in the registry
func (api *RegistryAPI) GetRepositories(ctx context.Context) ([]string, error) {
	endpoint := api.baseURL + "_catalog"
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	if api.username != "" {
		req.SetBasicAuth(api.username, api.password)
	}
	
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get catalog: status %d", resp.StatusCode)
	}
	
	var catalog catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}
	
	return catalog.Repositories, nil
}

// GetTags returns all tags for a repository
func (api *RegistryAPI) GetTags(ctx context.Context, repository string) ([]string, error) {
	endpoint := fmt.Sprintf("%s%s/tags/list", api.baseURL, repository)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	if api.username != "" {
		req.SetBasicAuth(api.username, api.password)
	}
	
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tags: status %d", resp.StatusCode)
	}
	
	var tags tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}
	
	return tags.Tags, nil
}

// GetManifest returns the manifest for a specific image tag
func (api *RegistryAPI) GetManifest(ctx context.Context, repository, tag string) (*manifestResponse, error) {
	endpoint := fmt.Sprintf("%s%s/manifests/%s", api.baseURL, repository, tag)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	// Accept v2 schema2 manifests
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	
	if api.username != "" {
		req.SetBasicAuth(api.username, api.password)
	}
	
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manifest: status %d", resp.StatusCode)
	}
	
	var manifest manifestResponse
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	
	return &manifest, nil
}

// GetImageInfo returns details about an image
func (api *RegistryAPI) GetImageInfo(ctx context.Context, repository string, tags []string) (dto.RegistryImageInfo, error) {
	if len(tags) == 0 {
		return dto.RegistryImageInfo{
			Name: repository,
			Tags: []string{},
			Size: 0,
			CreatedAt: time.Now(), // Default to now if we can't determine
		}, nil
	}
	
	// Get manifest for the first tag to determine size
	manifest, err := api.GetManifest(ctx, repository, tags[0])
	if err != nil {
		// Return partial info if we can't get manifest
		return dto.RegistryImageInfo{
			Name: repository,
			Tags: tags,
			Size: 0,
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
		CreatedAt: time.Now(), // Registry API doesn't provide creation time directly
	}, nil
}

// ListImages returns a list of all images in the registry with their details
func (api *RegistryAPI) ListImages(ctx context.Context) ([]dto.RegistryImageInfo, error) {
	// Get list of repositories
	repositories, err := api.GetRepositories(ctx)
	if err != nil {
		return nil, err
	}
	
	var images []dto.RegistryImageInfo
	
	// Get details for each repository
	for _, repo := range repositories {
		// Get tags for this repository
		tags, err := api.GetTags(ctx, repo)
		if err != nil {
			// Continue to next repo if we can't get tags
			continue
		}
		
		// Get image info
		imageInfo, err := api.GetImageInfo(ctx, repo, tags)
		if err != nil {
			// Add with partial info if we can't get full details
			images = append(images, dto.RegistryImageInfo{
				Name: repo,
				Tags: tags,
			})
			continue
		}
		
		images = append(images, imageInfo)
	}
	
	return images, nil
}

// ParseRegistryURL ensures the URL is properly formatted for API calls
func ParseRegistryURL(registryURL string) (string, error) {
	// Check if this is a Kubernetes cluster-internal URL (containing .svc.cluster.local)
	if strings.Contains(registryURL, ".svc.cluster.local") {
		// For Kubernetes internal registry URLs, we need to use the K8S_PROXY_URL environment variable
		// This proxy URL should be set in the application configuration
		proxyURL := os.Getenv("K8S_PROXY_URL")
		if proxyURL == "" {
			proxyURL = "https://kubernetes-proxy-millab.isacitra.com" // Fallback to default
		}

		// Extract the registry name from the URL (registry-<id>)
		parts := strings.Split(registryURL, ".")
		if len(parts) == 0 {
			return "", fmt.Errorf("invalid registry URL format")
		}

		// Construct URL to access registry through the Kubernetes proxy
		// Format: <proxy-url>/api/v1/namespaces/registry/services/<registry-name>:5000/proxy/v2/
		registryName := parts[0]
		portIdx := strings.LastIndex(registryName, ":")
		port := "5000" // Default port
		if portIdx > 0 {
			port = registryName[portIdx+1:]
			registryName = registryName[:portIdx]
		}

		proxyAccessURL := fmt.Sprintf("%s/api/v1/namespaces/registry/services/%s:%s/proxy/v2/", 
			proxyURL, registryName, port)
		return proxyAccessURL, nil
	}

	// Standard external URL processing
	// Add protocol if missing
	if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
		registryURL = "https://" + registryURL
	}
	
	// Parse and validate URL
	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return "", err
	}
	
	// Ensure path ends with /v2/
	if !strings.HasSuffix(parsedURL.Path, "/v2/") {
		if !strings.HasSuffix(parsedURL.Path, "/") {
			parsedURL.Path += "/"
		}
		if !strings.HasSuffix(parsedURL.Path, "/v2/") {
			parsedURL.Path += "v2/"
		}
	}
	
	return parsedURL.String(), nil
}
