package utils

import (
	"fmt"
)

// getRegistryResourceName returns a consistent name for Kubernetes resources
func GetRegistryResourceName(registryID string) string {
	// Use the full ID as requested to preserve ID for tracking
	return fmt.Sprintf("registry-%s", registryID)
}
