package utils

import (
	"fmt"
	"strings"

	"github.com/pendeploy-simple/models"
)

// generateNixpacksEnvFlags converts environment variables to nixpacks generate flags
func generateNixpacksEnvFlags(envVars models.EnvVars) string {
    if len(envVars) == 0 {
        return ""
    }
    
    var envFlags strings.Builder
    for key, value := range envVars {
        quotedValue := strings.ReplaceAll(value, "'", "'\"'\"'")
        envFlags.WriteString(fmt.Sprintf("--env %s='%s' ", key, quotedValue))
    }
    
    return envFlags.String()
}

// generateNixpacksConfigFlags converts service commands to nixpacks generate flags
func generateNixpacksConfigFlags(service models.Service) string {
    var configFlags strings.Builder
    
    if service.BuildCommand != "" {
        quotedValue := strings.ReplaceAll(service.BuildCommand, "'", "'\"'\"'")
        configFlags.WriteString(fmt.Sprintf("--build-cmd '%s' ", quotedValue))
    }
    
    if service.StartCommand != "" {
        quotedValue := strings.ReplaceAll(service.StartCommand, "'", "'\"'\"'")
        configFlags.WriteString(fmt.Sprintf("--start-cmd '%s' ", quotedValue))
    }
    
    return configFlags.String()
}