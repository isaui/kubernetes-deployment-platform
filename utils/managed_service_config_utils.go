// utils/managed_service_utils.go
package utils

import (
	"fmt"
	"github.com/pendeploy-simple/models"
)

// ManagedServiceConfig holds configuration for a managed service type
type ManagedServiceConfig struct {
	Port            int
	RequiresStorage bool
	DefaultVersion  string
	ServiceType     string // "StatefulSet" or "Deployment"
}

// GetManagedServiceConfigs returns configuration for all supported managed services
func GetManagedServiceConfigs() map[string]ManagedServiceConfig {
	return map[string]ManagedServiceConfig{
		"postgresql": {
			Port:            5432,
			RequiresStorage: true,
			DefaultVersion:  "15",
			ServiceType:     "StatefulSet",
		},
		"mysql": {
			Port:            3306,
			RequiresStorage: true,
			DefaultVersion:  "8.0",
			ServiceType:     "StatefulSet",
		},
		"redis": {
			Port:            6379,
			RequiresStorage: true,
			DefaultVersion:  "7",
			ServiceType:     "StatefulSet",
		},
		"mongodb": {
			Port:            27017,
			RequiresStorage: true,
			DefaultVersion:  "7.0",
			ServiceType:     "StatefulSet",
		},
		"minio": {
			Port:            9000,
			RequiresStorage: true,
			DefaultVersion:  "latest",
			ServiceType:     "StatefulSet",
		},
		"rabbitmq": {
			Port:            5672,
			RequiresStorage: true,
			DefaultVersion:  "3.12",
			ServiceType:     "StatefulSet",
		},
	}
}

// GetManagedServicePort returns the port for a managed service type
func GetManagedServicePort(managedType string) int {
	configs := GetManagedServiceConfigs()
	if config, exists := configs[managedType]; exists {
		return config.Port
	}
	return 8080 // fallback
}

// IsValidManagedServiceType checks if the managed service type is supported
func IsValidManagedServiceType(managedType string) bool {
	configs := GetManagedServiceConfigs()
	_, exists := configs[managedType]
	return exists
}

// GetManagedServiceDefaultVersion returns default version for a managed service type
func GetManagedServiceDefaultVersion(managedType string) string {
	configs := GetManagedServiceConfigs()
	if config, exists := configs[managedType]; exists {
		return config.DefaultVersion
	}
	return "latest"
}

// RequiresPersistentStorage checks if managed service needs storage
func RequiresPersistentStorage(managedType string) bool {
	configs := GetManagedServiceConfigs()
	if config, exists := configs[managedType]; exists {
		return config.RequiresStorage
	}
	return false
}

// GetManagedServiceType returns K8s resource type (StatefulSet/Deployment)
func GetManagedServiceType(managedType string) string {
	configs := GetManagedServiceConfigs()
	if config, exists := configs[managedType]; exists {
		return config.ServiceType
	}
	return "Deployment"
}

// GenerateManagedServiceEnvVars creates environment variables for managed service
func GenerateManagedServiceEnvVars(service models.Service) models.EnvVars {
	envVars := make(models.EnvVars)
	
	// Generate internal service hostname
	internalHost := fmt.Sprintf("%s.%s.svc.cluster.local", GetResourceName(service), service.EnvironmentID)
	
	// Generate external domain
	externalHost := GetManagedServiceExternalDomain(service)
	
	// Common vars for all services
	envVars["SERVICE_HOST"] = internalHost
	envVars["SERVICE_PORT"] = fmt.Sprintf("%d", service.Port)
	envVars["SERVICE_EXTERNAL_URL"] = fmt.Sprintf("https://%s", externalHost)
	
	// Service-specific environment variables
	switch service.ManagedType {
	case "postgresql":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)
		
		envVars["POSTGRES_DB"] = dbName
		envVars["POSTGRES_USER"] = dbUser
		envVars["POSTGRES_PASSWORD"] = dbPassword
		
		// Connection strings
		envVars["DATABASE_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["DATABASE_EXTERNAL_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:443/%s", dbUser, dbPassword, externalHost, dbName)
		
	case "mysql":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)
		
		envVars["MYSQL_DATABASE"] = dbName
		envVars["MYSQL_USER"] = dbUser
		envVars["MYSQL_PASSWORD"] = dbPassword
		envVars["MYSQL_ROOT_PASSWORD"] = GenerateSecurePassword(20)
		
		// Connection strings
		envVars["DATABASE_URL"] = fmt.Sprintf("mysql://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["DATABASE_EXTERNAL_URL"] = fmt.Sprintf("mysql://%s:%s@%s:443/%s", dbUser, dbPassword, externalHost, dbName)
		
	case "redis":
		redisPassword := GenerateSecurePassword(16)
		
		envVars["REDIS_PASSWORD"] = redisPassword
		
		// Connection strings
		envVars["REDIS_URL"] = fmt.Sprintf("redis://:%s@%s:%d", redisPassword, internalHost, service.Port)
		envVars["REDIS_EXTERNAL_URL"] = fmt.Sprintf("redis://:%s@%s:443", redisPassword, externalHost)
		
	case "mongodb":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)
		
		envVars["MONGO_INITDB_DATABASE"] = dbName
		envVars["MONGO_INITDB_ROOT_USERNAME"] = dbUser
		envVars["MONGO_INITDB_ROOT_PASSWORD"] = dbPassword
		
		// Connection strings
		envVars["MONGODB_URL"] = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["MONGODB_EXTERNAL_URL"] = fmt.Sprintf("mongodb://%s:%s@%s:443/%s", dbUser, dbPassword, externalHost, dbName)
		
	case "minio":
		accessKey := GenerateSecureID("access")
		secretKey := GenerateSecurePassword(20)
		
		envVars["MINIO_ROOT_USER"] = accessKey
		envVars["MINIO_ROOT_PASSWORD"] = secretKey
		
		// Connection info
		envVars["MINIO_ENDPOINT"] = internalHost
		envVars["MINIO_EXTERNAL_ENDPOINT"] = externalHost
		envVars["MINIO_ACCESS_KEY"] = accessKey
		envVars["MINIO_SECRET_KEY"] = secretKey
		
	case "rabbitmq":
		username := GenerateSecureID("user")
		password := GenerateSecurePassword(16)
		
		envVars["RABBITMQ_DEFAULT_USER"] = username
		envVars["RABBITMQ_DEFAULT_PASS"] = password
		
		// Connection strings
		envVars["RABBITMQ_URL"] = fmt.Sprintf("amqp://%s:%s@%s:%d", username, password, internalHost, service.Port)
		envVars["RABBITMQ_EXTERNAL_URL"] = fmt.Sprintf("amqp://%s:%s@%s:443", username, password, externalHost)
	}
	
	return envVars
}

// GetManagedServiceExternalDomain generates external domain for managed service
func GetManagedServiceExternalDomain(service models.Service) string {
	// Format: <service-name>-<env-short>.managed.app.isacitra.com
	serviceName := SanitizeLabel(service.Name)
	
	// Truncate environmentID to 6 characters
	shortEnvID := service.EnvironmentID
	if len(shortEnvID) > 6 {
		shortEnvID = shortEnvID[:6]
	}
	
	return fmt.Sprintf("%s-%s.managed.app.isacitra.com", serviceName, shortEnvID)
}

// GenerateSecureID generates a secure random ID for usernames/database names
func GenerateSecureID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, GenerateShortID())
}