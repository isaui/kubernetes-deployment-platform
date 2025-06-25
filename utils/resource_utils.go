package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseKubeResourceQuantity converts a Kubernetes resource quantity string to bytes
// Handles formats like: "100Mi", "2Gi", "4Ti", "500m", etc.
func ParseKubeResourceQuantity(quantity string) (int64, error) {
	// Handle CPU millicores (e.g. "500m")
	if strings.HasSuffix(quantity, "m") {
		value, err := strconv.ParseInt(strings.TrimSuffix(quantity, "m"), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse CPU millicores: %w", err)
		}
		return value * 1000000, nil // Convert millicores to nanocores
	}

	// Handle binary memory units
	memUnits := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"Pi": 1024 * 1024 * 1024 * 1024 * 1024,
		"Ei": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range memUnits {
		if strings.HasSuffix(quantity, suffix) {
			valueStr := strings.TrimSuffix(quantity, suffix)
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse memory quantity: %w", err)
			}
			return value * multiplier, nil
		}
	}

	// Handle decimal memory units
	decUnits := map[string]int64{
		"k":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
		"P":  1000 * 1000 * 1000 * 1000 * 1000,
		"E":  1000 * 1000 * 1000 * 1000 * 1000 * 1000,
		"K":  1000, // Also accept capital K
		"KB": 1000,
		"MB": 1000 * 1000,
		"GB": 1000 * 1000 * 1000,
		"TB": 1000 * 1000 * 1000 * 1000,
		"PB": 1000 * 1000 * 1000 * 1000 * 1000,
		"EB": 1000 * 1000 * 1000 * 1000 * 1000 * 1000,
	}

	for suffix, multiplier := range decUnits {
		if strings.HasSuffix(quantity, suffix) {
			valueStr := strings.TrimSuffix(quantity, suffix)
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse memory quantity: %w", err)
			}
			return value * multiplier, nil
		}
	}

	// Try to parse as plain number (bytes)
	value, err := strconv.ParseInt(quantity, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse quantity: %w", err)
	}
	return value, nil
}
