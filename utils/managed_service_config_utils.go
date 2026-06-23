// utils/managed_service_utils.go
package utils

import (
	"fmt"
	"github.com/pendeploy-simple/models"
	"strings"
)

// Port allocation ranges for managed services
const (
	MAX_PORT_ATTEMPTS = 100 // Maximum attempts to find available port
	PORT_RANGE_SIZE   = 100 // Default range size per service type
)

// ManagedServiceConfig holds configuration for a managed service type
type ManagedServiceConfig struct {
	Port            int
	RequiresStorage bool
	DefaultVersion  string
	ServiceType     string // "StatefulSet" or "Deployment"
	ExposureType    string // "TCPProxy" or "Ingress"
}

// ServiceExposureConfig defines how a service should be exposed
type ServiceExposureConfig struct {
	Name         string // "primary", "console", "management", etc.
	Port         int
	IsHTTP       bool // true for HTTP services, false for TCP
	Description  string
	Path         string // URL path for HTTP services
	ExposureType string // "TCPProxy" or "Ingress"
}

// GetManagedServiceConfigs returns configuration for all supported managed services
func GetManagedServiceConfigs() map[string]ManagedServiceConfig {
	return map[string]ManagedServiceConfig{
		"postgresql": {
			Port:            5432,
			RequiresStorage: true,
			DefaultVersion:  "15",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
		"mysql": {
			Port:            3306,
			RequiresStorage: true,
			DefaultVersion:  "8.0",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
		"redis": {
			Port:            6379,
			RequiresStorage: true,
			DefaultVersion:  "7",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
		"mongodb": {
			Port:            27017,
			RequiresStorage: true,
			DefaultVersion:  "7.0",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
		"minio": {
			Port:            9000,
			RequiresStorage: true,
			DefaultVersion:  "latest",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
		"rabbitmq": {
			Port:            5672,
			RequiresStorage: true,
			DefaultVersion:  "3.12",
			ServiceType:     "StatefulSet",
			ExposureType:    "TCPProxy",
		},
	}
}

// GetManagedServiceExposureConfig returns all exposure configurations for a service type
func GetManagedServiceExposureConfig(managedType string) []ServiceExposureConfig {
	switch managedType {
	case "postgresql":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         5432,
				IsHTTP:       false,
				Description:  "PostgreSQL Database Connection",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
		}
	case "mysql":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         3306,
				IsHTTP:       false,
				Description:  "MySQL Database Connection",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
		}
	case "redis":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         6379,
				IsHTTP:       false,
				Description:  "Redis Database Connection",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
		}
	case "mongodb":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         27017,
				IsHTTP:       false,
				Description:  "MongoDB Database Connection",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
		}
	case "minio":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         9000,
				IsHTTP:       false,
				Description:  "MinIO S3 API",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
			{
				Name:         "console",
				Port:         9001,
				IsHTTP:       true,
				Description:  "MinIO Console (Web UI)",
				Path:         "/",
				ExposureType: "Ingress",
			},
		}
	case "rabbitmq":
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         5672,
				IsHTTP:       false,
				Description:  "RabbitMQ AMQP Connection",
				Path:         "/",
				ExposureType: "TCPProxy",
			},
			{
				Name:         "management",
				Port:         15672,
				IsHTTP:       true,
				Description:  "RabbitMQ Management UI",
				Path:         "/",
				ExposureType: "Ingress",
			},
		}
	default:
		return []ServiceExposureConfig{
			{
				Name:         "primary",
				Port:         8080,
				IsHTTP:       true,
				Description:  "Default HTTP Service",
				Path:         "/",
				ExposureType: "Ingress",
			},
		}
	}
}

// GetManagedServicePort returns the primary port for a managed service type
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

// GenerateManagedServiceEnvVars creates comprehensive environment variables for managed services.
func GenerateManagedServiceEnvVars(service models.Service, externalHost string, externalPort int) models.EnvVars {
	envVars := make(models.EnvVars)

	// Generate internal service hostname (primary service)
	internalHost := fmt.Sprintf("%s.%s.svc.cluster.local", GetResourceName(service), service.EnvironmentID)

	// Get all exposure configs for this service
	exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)

	// Common vars for all services
	envVars["SERVICE_HOST"] = internalHost
	envVars["SERVICE_PORT"] = fmt.Sprintf("%d", service.Port)

	// Generate connection info for all exposed services
	for _, config := range exposureConfigs {
		if config.Name == "primary" {
			// Primary service gets main variables through the shared TCP proxy.
			if config.ExposureType == "TCPProxy" {
				envVars["SERVICE_EXTERNAL_HOST"] = externalHost
				envVars["SERVICE_EXTERNAL_PORT"] = fmt.Sprintf("%d", externalPort)
			} else {
				// HTTP services use domain
				externalHost := GetManagedServiceExternalDomain(service, config.Name)
				envVars["SERVICE_EXTERNAL_URL"] = fmt.Sprintf("https://%s", externalHost)
			}
		} else {
			// Secondary services get prefixed variables
			prefix := strings.ToUpper(config.Name)
			if config.ExposureType == "TCPProxy" {
				// TCP services (if any secondary TCP services in future)
				envVars[fmt.Sprintf("%s_EXTERNAL_HOST", prefix)] = externalHost
				envVars[fmt.Sprintf("%s_EXTERNAL_PORT", prefix)] = fmt.Sprintf("%d", config.Port) // Would need separate allocation
			} else {
				// HTTP services use domain
				externalHost := GetManagedServiceExternalDomain(service, config.Name)
				envVars[fmt.Sprintf("%s_EXTERNAL_URL", prefix)] = fmt.Sprintf("https://%s", externalHost)
			}

			// Internal URLs
			internalServiceHost := fmt.Sprintf("%s-%s.%s.svc.cluster.local", GetResourceName(service), config.Name, service.EnvironmentID)
			envVars[fmt.Sprintf("%s_HOST", prefix)] = internalServiceHost
			envVars[fmt.Sprintf("%s_PORT", prefix)] = fmt.Sprintf("%d", config.Port)
		}
	}

	// Service-specific environment variables with TCP proxy connections
	switch service.ManagedType {
	case "postgresql":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)

		envVars["POSTGRES_DB"] = dbName
		envVars["POSTGRES_USER"] = dbUser
		envVars["POSTGRES_PASSWORD"] = dbPassword

		// Connection strings - use internal DNS and the shared TCP proxy for external access.
		envVars["DATABASE_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["DATABASE_EXTERNAL_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPassword, externalHost, externalPort, dbName)

	case "mysql":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)

		envVars["MYSQL_DATABASE"] = dbName
		envVars["MYSQL_USER"] = dbUser
		envVars["MYSQL_PASSWORD"] = dbPassword
		envVars["MYSQL_ROOT_PASSWORD"] = GenerateSecurePassword(20)

		// Connection strings - use internal DNS and the shared TCP proxy for external access.
		envVars["DATABASE_URL"] = fmt.Sprintf("mysql://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["DATABASE_EXTERNAL_URL"] = fmt.Sprintf("mysql://%s:%s@%s:%d/%s", dbUser, dbPassword, externalHost, externalPort, dbName)

	case "redis":
		redisPassword := GenerateSecurePassword(16)

		envVars["REDIS_PASSWORD"] = redisPassword

		// Connection strings - use internal DNS and the shared TCP proxy for external access.
		envVars["REDIS_URL"] = fmt.Sprintf("redis://:%s@%s:%d", redisPassword, internalHost, service.Port)
		envVars["REDIS_EXTERNAL_URL"] = fmt.Sprintf("redis://:%s@%s:%d", redisPassword, externalHost, externalPort)

	case "mongodb":
		dbName := GenerateSecureID("db")
		dbUser := GenerateSecureID("user")
		dbPassword := GenerateSecurePassword(16)

		envVars["MONGO_INITDB_DATABASE"] = dbName
		envVars["MONGO_INITDB_ROOT_USERNAME"] = dbUser
		envVars["MONGO_INITDB_ROOT_PASSWORD"] = dbPassword

		// Connection strings - use internal DNS and the shared TCP proxy for external access.
		envVars["MONGODB_URL"] = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", dbUser, dbPassword, internalHost, service.Port, dbName)
		envVars["MONGODB_EXTERNAL_URL"] = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", dbUser, dbPassword, externalHost, externalPort, dbName)

	case "minio":
		accessKey := GenerateSecureID("access")
		secretKey := GenerateSecurePassword(20)

		// MinIO requires specific environment variables
		envVars["MINIO_ROOT_USER"] = accessKey
		envVars["MINIO_ROOT_PASSWORD"] = secretKey
		envVars["MINIO_CONSOLE_ADDRESS"] = ":9001" // Console listens on port 9001
		envVars["MINIO_BROWSER"] = "on"            // Enable browser/console

		// Console URL configuration
		consoleHost := GetManagedServiceExternalDomain(service, "console")
		envVars["MINIO_SERVER_URL"] = fmt.Sprintf("http://%s:9000", internalHost)      // Tell console where API server is
		envVars["MINIO_BROWSER_REDIRECT_URL"] = fmt.Sprintf("https://%s", consoleHost) // Console redirect URL

		// Connection info - API via TCP proxy, Console via domain
		envVars["MINIO_ENDPOINT"] = fmt.Sprintf("%s:%d", internalHost, service.Port)
		envVars["MINIO_EXTERNAL_ENDPOINT"] = fmt.Sprintf("%s:%d", externalHost, externalPort)
		envVars["MINIO_ACCESS_KEY"] = accessKey
		envVars["MINIO_SECRET_KEY"] = secretKey

		// S3 API via TCP proxy, Console via domain
		envVars["S3_API_URL"] = fmt.Sprintf("http://%s:%d", externalHost, externalPort)
		envVars["MINIO_CONSOLE_URL"] = fmt.Sprintf("https://%s", consoleHost)

	case "rabbitmq":
		username := GenerateSecureID("user")
		password := GenerateSecurePassword(16)

		envVars["RABBITMQ_DEFAULT_USER"] = username
		envVars["RABBITMQ_DEFAULT_PASS"] = password

		// Connection info - AMQP via TCP proxy, Management via domain
		mgmtHost := GetManagedServiceExternalDomain(service, "management")

		// AMQP Connection strings - use internal DNS and the shared TCP proxy for external access.
		envVars["RABBITMQ_URL"] = fmt.Sprintf("amqp://%s:%s@%s:%d", username, password, internalHost, service.Port)
		envVars["RABBITMQ_EXTERNAL_URL"] = fmt.Sprintf("amqp://%s:%s@%s:%d", username, password, externalHost, externalPort)

		// Management UI - HTTP service uses domain
		envVars["RABBITMQ_MANAGEMENT_URL"] = fmt.Sprintf("https://%s", mgmtHost)
	}

	return envVars
}

// GetManagedServiceExternalDomain generates external domain for HTTP services only
func GetManagedServiceExternalDomain(service models.Service, endpointName ...string) string {
	serviceName := SanitizeLabel(service.Name)

	// Truncate environmentID to 6 characters
	shortEnvID := service.EnvironmentID
	if len(shortEnvID) > 6 {
		shortEnvID = shortEnvID[:6]
	}

	// Default to primary endpoint if no name specified
	endpoint := "primary"
	if len(endpointName) > 0 && endpointName[0] != "" {
		endpoint = endpointName[0]
	}

	if endpoint == "primary" {
		// Primary endpoint gets simple domain
		return fmt.Sprintf("%s-%s.managed.%s", serviceName, shortEnvID, GetDefaultDomain())
	} else {
		// Secondary endpoints get prefixed domain
		return fmt.Sprintf("%s-%s-%s.managed.%s", serviceName, endpoint, shortEnvID, GetDefaultDomain())
	}
}

// GetAllManagedServiceDomains returns all external domains for a managed service
func GetAllManagedServiceDomains(service models.Service) map[string]string {
	domains := make(map[string]string)
	exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)

	for _, config := range exposureConfigs {
		if config.ExposureType == "Ingress" {
			domains[config.Name] = GetManagedServiceExternalDomain(service, config.Name)
		}
	}

	return domains
}

// GetManagedServiceConnectionInfo returns comprehensive connection information.
func GetManagedServiceConnectionInfo(service models.Service, externalHost string, externalPort int) map[string]interface{} {
	info := make(map[string]interface{})
	exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)

	// Service metadata
	info["name"] = service.Name
	info["type"] = service.ManagedType
	info["status"] = service.Status

	// Endpoints
	endpoints := make(map[string]map[string]string)
	for _, config := range exposureConfigs {
		endpoint := make(map[string]string)
		endpoint["description"] = config.Description
		endpoint["port"] = fmt.Sprintf("%d", config.Port)
		endpoint["protocol"] = "TCP"

		if config.ExposureType == "TCPProxy" {
			endpoint["protocol"] = "TCP"
			endpoint["external_host"] = externalHost
			if config.Name == "primary" {
				endpoint["external_port"] = fmt.Sprintf("%d", externalPort)
			} else {
				endpoint["external_port"] = fmt.Sprintf("%d", config.Port) // Would need separate allocation
			}
		} else {
			endpoint["protocol"] = "HTTP"
			endpoint["external_url"] = fmt.Sprintf("https://%s", GetManagedServiceExternalDomain(service, config.Name))
		}

		// Internal connection
		if config.Name == "primary" {
			endpoint["internal_host"] = fmt.Sprintf("%s.%s.svc.cluster.local", GetResourceName(service), service.EnvironmentID)
		} else {
			endpoint["internal_host"] = fmt.Sprintf("%s-%s.%s.svc.cluster.local", GetResourceName(service), config.Name, service.EnvironmentID)
		}
		endpoint["internal_port"] = fmt.Sprintf("%d", config.Port)

		endpoints[config.Name] = endpoint
	}
	info["endpoints"] = endpoints

	// Connection credentials (if applicable)
	if credentials := extractCredentialsFromEnvVars(service.EnvVars, service.ManagedType); len(credentials) > 0 {
		info["credentials"] = credentials
	}

	return info
}

// extractCredentialsFromEnvVars extracts connection credentials from environment variables
func extractCredentialsFromEnvVars(envVars models.EnvVars, managedType string) map[string]string {
	credentials := make(map[string]string)

	switch managedType {
	case "postgresql":
		if user, exists := envVars["POSTGRES_USER"]; exists {
			credentials["username"] = user
		}
		if pass, exists := envVars["POSTGRES_PASSWORD"]; exists {
			credentials["password"] = pass
		}
		if db, exists := envVars["POSTGRES_DB"]; exists {
			credentials["database"] = db
		}
		if url, exists := envVars["DATABASE_URL"]; exists {
			credentials["connection_string"] = url
		}
		if externalUrl, exists := envVars["DATABASE_EXTERNAL_URL"]; exists {
			credentials["external_connection_string"] = externalUrl
		}

	case "mysql":
		if user, exists := envVars["MYSQL_USER"]; exists {
			credentials["username"] = user
		}
		if pass, exists := envVars["MYSQL_PASSWORD"]; exists {
			credentials["password"] = pass
		}
		if db, exists := envVars["MYSQL_DATABASE"]; exists {
			credentials["database"] = db
		}
		if url, exists := envVars["DATABASE_URL"]; exists {
			credentials["connection_string"] = url
		}
		if externalUrl, exists := envVars["DATABASE_EXTERNAL_URL"]; exists {
			credentials["external_connection_string"] = externalUrl
		}

	case "redis":
		if pass, exists := envVars["REDIS_PASSWORD"]; exists {
			credentials["password"] = pass
		}
		if url, exists := envVars["REDIS_URL"]; exists {
			credentials["connection_string"] = url
		}
		if externalUrl, exists := envVars["REDIS_EXTERNAL_URL"]; exists {
			credentials["external_connection_string"] = externalUrl
		}

	case "mongodb":
		if user, exists := envVars["MONGO_INITDB_ROOT_USERNAME"]; exists {
			credentials["username"] = user
		}
		if pass, exists := envVars["MONGO_INITDB_ROOT_PASSWORD"]; exists {
			credentials["password"] = pass
		}
		if db, exists := envVars["MONGO_INITDB_DATABASE"]; exists {
			credentials["database"] = db
		}
		if url, exists := envVars["MONGODB_URL"]; exists {
			credentials["connection_string"] = url
		}
		if externalUrl, exists := envVars["MONGODB_EXTERNAL_URL"]; exists {
			credentials["external_connection_string"] = externalUrl
		}

	case "minio":
		if accessKey, exists := envVars["MINIO_ACCESS_KEY"]; exists {
			credentials["access_key"] = accessKey
		}
		if secretKey, exists := envVars["MINIO_SECRET_KEY"]; exists {
			credentials["secret_key"] = secretKey
		}
		if endpoint, exists := envVars["S3_API_URL"]; exists {
			credentials["s3_endpoint"] = endpoint
		}

	case "rabbitmq":
		if user, exists := envVars["RABBITMQ_DEFAULT_USER"]; exists {
			credentials["username"] = user
		}
		if pass, exists := envVars["RABBITMQ_DEFAULT_PASS"]; exists {
			credentials["password"] = pass
		}
		if url, exists := envVars["RABBITMQ_URL"]; exists {
			credentials["connection_string"] = url
		}
		if externalUrl, exists := envVars["RABBITMQ_EXTERNAL_URL"]; exists {
			credentials["external_connection_string"] = externalUrl
		}
	}

	return credentials
}

// GenerateSecureID generates a secure random ID for usernames/database names
func GenerateSecureID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, GenerateShortID())
}
