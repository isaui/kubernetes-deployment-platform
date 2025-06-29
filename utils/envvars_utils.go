package utils

import (
	"github.com/pendeploy-simple/models"
	corev1 "k8s.io/api/core/v1"
)

// Helper function to convert environment variables map to Kubernetes EnvVar slice
func createEnvVarsFromMap(envVars models.EnvVars) []corev1.EnvVar {
	if len(envVars) == 0 {
		return nil
	}

	result := make([]corev1.EnvVar, 0, len(envVars))
	for key, value := range envVars {
		result = append(result, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
	
	return result
}