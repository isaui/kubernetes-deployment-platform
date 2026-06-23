package utils

import (
	"errors"
	"net/url"
	"strings"
)

// CleanRegistryURL returns the canonical registry address used in image tags.
// Container image names must not include a URL scheme.
func CleanRegistryURL(registryURL string) string {
	// Remove protocol prefix
	cleaned := strings.TrimPrefix(registryURL, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")

	// Remove any leading slashes that might have been added
	cleaned = strings.TrimPrefix(cleaned, "/")

	return cleaned
}

// GetRegistryURLWithProtocol returns an HTTPS URL for HTTP clients.
func GetRegistryURLWithProtocol(registryURL string) string {
	return "https://" + CleanRegistryURL(registryURL)
}

func GetRegistryURLForKaniko(registryURL string) string {
	if IsInsecureRegistry(registryURL) {
		if strings.HasPrefix(registryURL, "http://") {
			return registryURL
		}
		return "http://" + CleanRegistryURL(registryURL)
	}
	return GetRegistryURLWithProtocol(registryURL)
}

func IsInsecureRegistry(registryURL string) bool {
	parsed, err := parseRegistryURL(registryURL)
	if err != nil {
		return false
	}

	if parsed.Scheme == "http" {
		return true
	}

	host := strings.ToLower(parsed.Hostname())
	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasSuffix(host, ".svc.cluster.local") ||
		strings.HasSuffix(host, ".local")
}

func KanikoRegistryArgs(registryURL string) []string {
	if IsInsecureRegistry(registryURL) {
		return []string{"--insecure", "--skip-tls-verify"}
	}
	return nil
}

func parseRegistryURL(registryURL string) (*url.URL, error) {
	normalized := strings.TrimSpace(registryURL)
	if !strings.Contains(normalized, "://") {
		normalized = "registry://" + normalized
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, err
	}

	if parsed.Host == "" && parsed.Path != "" {
		parsed.Host = parsed.Path
		parsed.Path = ""
	}

	if parsed.Hostname() == "" {
		return nil, errEmptyRegistryHost
	}

	return parsed, nil
}

var errEmptyRegistryHost = errors.New("empty registry host")
