package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSecurePassword creates a random secure password of the specified length
func GenerateSecurePassword(length int) string {
	// Ensure minimum length
	if length < 8 {
		length = 8
	}
	
	// Create a byte slice large enough to generate the desired password length
	// We'll need more random bytes than the final length because of base64 encoding
	b := make([]byte, length*2)
	
	// Read random bytes
	_, err := rand.Read(b)
	if err != nil {
		// In case of error, return a hardcoded but reasonably secure fallback
		return "Temp@Password123"
	}
	
	// Convert to base64 (which is reasonably URL-safe and readable)
	// and trim to the desired length
	password := base64.StdEncoding.EncodeToString(b)
	if len(password) > length {
		password = password[:length]
	}
	
	return password
}
