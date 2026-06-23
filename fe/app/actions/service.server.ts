// api/service.server.ts - UPDATED untuk managed services
import type { Service, CreateServicePayload } from '~/types/service';
import type { ServiceUpdate } from '~/types/service-update';
import { formatUpdatePayload, validateServiceUpdate } from '~/types/service-update';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Helper function to build headers with authentication
 */
function buildHeaders(request?: Request): HeadersInit {
  const headers: HeadersInit = {
    'Accept': 'application/json'
  };
  
  // Forward authorization cookie if request provided
  if (request?.headers) {
    const cookie = request.headers.get('Cookie');
    if (cookie) {
      headers['Cookie'] = cookie;
    }
  }
  
  return headers;
}

/**
 * Fetch list of services for the current user
 */
export async function getServices(request: Request): Promise<Service[]> {
  const response = await fetch(`${API_URL}/services`, {
    method: 'GET',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch services: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.services || [];
}

/**
 * Fetch list of services for a specific project
 */
export async function getProjectServices(projectId: string, request: Request): Promise<Service[]> {
  const response = await fetch(`${API_URL}/projects/${projectId}/services`, {
    method: 'GET',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch project services: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.services || [];
}

/**
 * Fetch a single service by ID
 */
export async function getService(id: string, request: Request): Promise<Service> {
  const response = await fetch(`${API_URL}/services/${id}`, {
    method: 'GET',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch service: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Create a new service - UPDATED untuk support managed services
 */
export async function createService(service: CreateServicePayload, request: Request): Promise<Service> {
  // Validate service data based on type
  if (service.type === 'git') {
    if (!service.repoUrl) {
      throw new Error('Repository URL is required for git services');
    }
  } else if (service.type === 'managed') {
    if (!service.managedType) {
      throw new Error('Managed service type is required for managed services');
    }
    
    // Validate that git-specific fields are not provided
    if (service.repoUrl || service.branch || service.buildCommand || service.startCommand || service.envVars) {
      throw new Error('Git-specific fields are not allowed for managed services');
    }
    
    // Port should not be specified for managed services
    if (service.port) {
      throw new Error('Port is auto-determined for managed services and cannot be specified');
    }
  }
  
  const response = await fetch(`${API_URL}/services`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      ...buildHeaders(request),
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(service)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to create service: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Update an existing service - UPDATED untuk use ServiceUpdate format
 */
export async function updateService(id: string, serviceUpdate: ServiceUpdate, request: Request): Promise<Service> {
  // Validate the update
  const validationErrors = validateServiceUpdate(serviceUpdate);
  if (validationErrors.length > 0) {
    throw new Error(`Validation failed: ${validationErrors.join(', ')}`);
  }
  
  // Convert to backend format
  const payload = formatUpdatePayload(serviceUpdate);
  
  const response = await fetch(`${API_URL}/services/${id}`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      ...buildHeaders(request),
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to update service: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Legacy update function for backward compatibility
 * @deprecated Use updateService with ServiceUpdate instead
 */
export async function updateServiceLegacy(id: string, service: Partial<Service>, request: Request): Promise<Service> {
  console.warn('updateServiceLegacy is deprecated. Use updateService with ServiceUpdate format.');
  
  const response = await fetch(`${API_URL}/services/${id}`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      ...buildHeaders(request),
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(service)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to update service: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Delete a service
 */
export async function deleteService(id: string, request: Request): Promise<void> {
  const response = await fetch(`${API_URL}/services/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to delete service: ${response.statusText}`);
  }
}

/**
 * Get latest deployment for a git service
 */
export async function getLatestDeployment(serviceId: string, request: Request) {
  const response = await fetch(`${API_URL}/services/${serviceId}/latest-deployment`, {
    method: 'GET',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to fetch latest deployment: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.deployment;
}

/**
 * Get deployment list for a git service
 */
export async function getDeploymentList(serviceId: string, request: Request) {
  const response = await fetch(`${API_URL}/services/${serviceId}/deployments`, {
    method: 'GET',
    credentials: 'include',
    headers: buildHeaders(request)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to fetch deployments: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.deployments;
}

/**
 * Helper function to check if a service supports deployments
 */
export function supportsDeployments(service: Service): boolean {
  return service.type === 'git';
}

/**
 * Helper function to get managed service connection information
 */
export function getManagedServiceConnection(service: Service): {
  internalUrl?: string;
  externalUrl?: string;
  connectionString?: string;
  credentials?: Record<string, string>;
} | null {
  if (service.type !== 'managed' || !service.envVars) {
    return null;
  }
  
  const envVars = service.envVars;
  const connection: any = {};
  
  // Internal URL
  if (envVars.SERVICE_HOST && envVars.SERVICE_PORT) {
    connection.internalUrl = `${envVars.SERVICE_HOST}:${envVars.SERVICE_PORT}`;
  }
  
  // External URL
  if (envVars.SERVICE_EXTERNAL_URL) {
    connection.externalUrl = envVars.SERVICE_EXTERNAL_URL;
  }
  
  // Connection strings
  if (envVars.DATABASE_URL) connection.connectionString = envVars.DATABASE_URL;
  else if (envVars.REDIS_URL) connection.connectionString = envVars.REDIS_URL;
  else if (envVars.MONGODB_URL) connection.connectionString = envVars.MONGODB_URL;
  else if (envVars.RABBITMQ_URL) connection.connectionString = envVars.RABBITMQ_URL;
  
  // Extract credentials
  const credentials: Record<string, string> = {};
  const credentialKeys = Object.keys(envVars).filter(key => 
    key.includes('USER') || key.includes('PASSWORD') || key.includes('ACCESS_KEY') || 
    key.includes('SECRET') || key.includes('DB') || key.includes('DATABASE')
  );
  
  credentialKeys.forEach(key => {
    credentials[key.toLowerCase()] = envVars[key];
  });
  
  if (Object.keys(credentials).length > 0) {
    connection.credentials = credentials;
  }
  
  return connection;
}