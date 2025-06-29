package utils

import (
	"strings"
)

// CleanRegistryURL removes protocol prefix and returns clean registry URL for kaniko
func CleanRegistryURL(registryURL string) string {
	// Remove protocol prefix
	cleaned := strings.TrimPrefix(registryURL, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")
	
	// Remove any leading slashes that might have been added
	cleaned = strings.TrimPrefix(cleaned, "/")
	
	return cleaned
}

// GetRegistryURLWithProtocol ensures registry URL has HTTPS protocol
func GetRegistryURLWithProtocol(registryURL string) string {
	if !strings.HasPrefix(registryURL, "https://") && !strings.HasPrefix(registryURL, "http://") {
		return "https://" + registryURL
	} else if strings.HasPrefix(registryURL, "http://") {
		return strings.Replace(registryURL, "http://", "https://", 1)
	}
	return registryURL
}