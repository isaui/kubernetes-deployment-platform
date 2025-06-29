package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// EncodeBasicAuth encodes username and password for HTTP Basic Authentication
// Returns base64 encoded string of "username:password"
func EncodeBasicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// DecodeBasicAuth decodes a base64 encoded basic auth string
// Returns username and password
func DecodeBasicAuth(encoded string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode base64: %v", err)
	}
	
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid basic auth format")
	}
	
	return parts[0], parts[1], nil
}

// GenerateShortID generates a short, URL-safe random ID
// Format: 8 characters, lowercase alphanumeric
// Example: "x7k9m2p1"
func GenerateShortID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8
	
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}
	
	return string(result)
}

// GenerateID generates a longer random ID
// Format: 16 characters, lowercase alphanumeric
// Example: "a7k9m2p1x5n8q3r6"
func GenerateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 16
	
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}
	
	return string(result)
}

// GenerateJobName generates a Kubernetes-compliant job name
// Format: prefix-shortid-timestamp
// Example: "build-x7k9m2p1-1640995200"
func GenerateJobName(prefix string) string {
	shortID := GenerateShortID()
	timestamp := time.Now().Unix()
	
	// Ensure Kubernetes naming compliance
	prefix = strings.ToLower(prefix)
	prefix = strings.ReplaceAll(prefix, "_", "-")
	
	return fmt.Sprintf("%s-%s-%d", prefix, shortID, timestamp)
}


// IsValidKubernetesName checks if a string is a valid Kubernetes resource name
func IsValidKubernetesName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	
	// Must start and end with alphanumeric
	if !isAlphanumeric(name[0]) || !isAlphanumeric(name[len(name)-1]) {
		return false
	}
	
	// Check each character
	for _, char := range name {
		if !isAlphanumeric(byte(char)) && char != '-' && char != '.' {
			return false
		}
	}
	
	return true
}

// isAlphanumeric checks if a byte is alphanumeric
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}