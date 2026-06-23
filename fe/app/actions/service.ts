// api/service.ts - UPDATED untuk managed services
import { Deployment } from "~/types/deployment";
import { Service, CreateServicePayload, isGitService } from "~/types/service";
import { ServiceUpdate, formatUpdatePayload, validateServiceUpdate } from "~/types/service-update";

// Function to get all services for the current user
export async function getServices(): Promise<Service[]> {
  const response = await fetch(
    `/api/v1/services`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    throw new Error("Failed to fetch services");
  }

  const result = await response.json();
  return result.data.services;
}

// Function to get a specific service
export async function getService(serviceId: string): Promise<Service> {
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    throw new Error("Failed to fetch service");
  }

  const result = await response.json();
  return result.data;
}

// Function to get latest deployment (hanya untuk git services)
export async function getLatestDeployment(serviceId: string): Promise<Deployment> {
  const response = await fetch(
    `/api/v1/services/${serviceId}/latest-deployment`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to fetch latest deployment");
  }

  const result = await response.json();
  return result.data.deployment;
}

// Function to get deployment list (hanya untuk git services)
export async function getDeploymentList(serviceId: string): Promise<Deployment[]> {
  const response = await fetch(
    `/api/v1/services/${serviceId}/deployments`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to fetch deployments");
  }

  const result = await response.json();
  return result.data.deployments;
}

// Function to create a new service - UPDATED untuk support managed services
export async function createService(data: CreateServicePayload): Promise<Service> {
  // Validate required fields based on service type
  if (data.type === 'git') {
    if (!data.repoUrl) {
      throw new Error("Repository URL is required for git services");
    }
  } else if (data.type === 'managed') {
    if (!data.managedType) {
      throw new Error("Managed service type is required for managed services");
    }
    
    // Ensure git-specific fields are not provided for managed services
    if (data.repoUrl || data.branch || data.buildCommand || data.startCommand || data.envVars) {
      throw new Error("Git-specific fields are not allowed for managed services");
    }
    
    // Port is auto-determined for managed services
    if (data.port) {
      throw new Error("Port is auto-determined for managed services");
    }
  }

  const response = await fetch(
    `/api/v1/services`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to create service");
  }

  const result = await response.json();
  return result.data;
}

// Function to update a service - UPDATED untuk use new ServiceUpdate format
export async function updateService(serviceId: string, update: ServiceUpdate): Promise<Service> {
  // Validate the update
  const validationErrors = validateServiceUpdate(update);
  if (validationErrors.length > 0) {
    throw new Error(`Validation failed: ${validationErrors.join(', ')}`);
  }
  
  // Convert to backend format
  const payload = formatUpdatePayload(update);
  
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to update service");
  }

  const result = await response.json();
  return result.data;
}

// Function to delete a service
export async function deleteService(serviceId: string): Promise<void> {
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "DELETE",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to delete service");
  }
}

// Helper function to check if service supports deployments
export function supportsDeployments(service: Service): boolean {
  return isGitService(service);
}

// Helper function to get service connection info (untuk managed services)
export function getServiceConnectionInfo(service: Service): {
  internalUrl?: string;
  externalUrl?: string;
  credentials?: Record<string, string>;
} | null {
  if (service.type !== 'managed' || !service.envVars) {
    return null;
  }
  
  const envVars = service.envVars;
  
  return {
    internalUrl: envVars.SERVICE_HOST ? `${envVars.SERVICE_HOST}:${envVars.SERVICE_PORT}` : undefined,
    externalUrl: envVars.SERVICE_EXTERNAL_URL,
    credentials: extractCredentials(service),
  };
}

// Helper function to extract credentials from managed service env vars
function extractCredentials(service: Service): Record<string, string> {
  if (service.type !== 'managed' || !service.envVars) {
    return {};
  }
  
  const envVars = service.envVars;
  const credentials: Record<string, string> = {};
  
  // Extract common credential patterns
  if (envVars.DATABASE_URL) credentials.databaseUrl = envVars.DATABASE_URL;
  if (envVars.REDIS_URL) credentials.redisUrl = envVars.REDIS_URL;
  if (envVars.MONGODB_URL) credentials.mongoUrl = envVars.MONGODB_URL;
  if (envVars.RABBITMQ_URL) credentials.rabbitmqUrl = envVars.RABBITMQ_URL;
  
  // Extract individual credentials
  const credentialKeys = [
    'POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB',
    'MYSQL_USER', 'MYSQL_PASSWORD', 'MYSQL_DATABASE',
    'REDIS_PASSWORD',
    'MONGO_INITDB_ROOT_USERNAME', 'MONGO_INITDB_ROOT_PASSWORD',
    'MINIO_ACCESS_KEY', 'MINIO_SECRET_KEY',
    'RABBITMQ_DEFAULT_USER', 'RABBITMQ_DEFAULT_PASS'
  ];
  
  credentialKeys.forEach(key => {
    if (envVars[key]) {
      credentials[key.toLowerCase()] = envVars[key];
    }
  });
  
  return credentials;
}