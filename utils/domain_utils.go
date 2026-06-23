package utils

import (
	"os"
	"strings"
)

const fallbackDefaultDomain = "app.isacitra.com"

func GetDefaultDomain() string {
	domain := strings.TrimSpace(os.Getenv("DEFAULT_DOMAIN"))
	if domain == "" {
		return fallbackDefaultDomain
	}

	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	return strings.Trim(domain, "/")
}
