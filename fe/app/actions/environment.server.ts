import type { Environment } from '~/types/project';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Fetch environments for a specific project
 * 
 * @param {string} projectId - The project ID to fetch environments for
 * @param {Request} request - Remix request with credentials
 * @returns List of environments
 */
export async function getEnvironments(projectId: string, request: Request): Promise<Environment[]> {
  // Build headers with authentication if request is provided
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
  
  const response = await fetch(`${API_URL}/projects/${projectId}/environments`, {
    method: 'GET',
    credentials: 'include',
    headers
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch environments: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.environments || [];
}

/**
 * Fetch a single environment by ID
 * 
 * @param {string} id - Environment ID to fetch
 * @param {Request} request - Remix request with credentials
 * @returns Environment details
 */
export async function getEnvironment(id: string, request: Request): Promise<Environment> {
  // Build headers with authentication if request is provided
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
  
  const response = await fetch(`${API_URL}/environments/${id}`, {
    method: 'GET',
    credentials: 'include',
    headers
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch environment: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Create a new environment
 * 
 * @param {Partial<Environment>} environment - Environment data to create
 * @param {Request} request - Remix request with credentials
 * @returns Newly created environment
 */
export async function createEnvironment(environment: Partial<Environment>, request: Request): Promise<Environment> {
  // Build headers with authentication if request is provided
  const headers: HeadersInit = {
    'Accept': 'application/json',
    'Content-Type': 'application/json'
  };
  
  // Forward authorization cookie if request provided
  if (request?.headers) {
    const cookie = request.headers.get('Cookie');
    if (cookie) {
      headers['Cookie'] = cookie;
    }
  }
  
  const response = await fetch(`${API_URL}/environments`, {
    method: 'POST',
    credentials: 'include',
    headers,
    body: JSON.stringify(environment)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to create environment: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Update an existing environment
 * 
 * @param {string} id - Environment ID to update
 * @param {Partial<Environment>} environment - Updated environment data
 * @param {Request} request - Remix request with credentials
 * @returns Updated environment
 */
export async function updateEnvironment(id: string, environment: Partial<Environment>, request: Request): Promise<Environment> {
  // Build headers with authentication if request is provided
  const headers: HeadersInit = {
    'Accept': 'application/json',
    'Content-Type': 'application/json'
  };
  
  // Forward authorization cookie if request provided
  if (request?.headers) {
    const cookie = request.headers.get('Cookie');
    if (cookie) {
      headers['Cookie'] = cookie;
    }
  }
  
  const response = await fetch(`${API_URL}/environments/${id}`, {
    method: 'PUT',
    credentials: 'include',
    headers,
    body: JSON.stringify(environment)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to update environment: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Delete an environment
 * 
 * @param {string} id - Environment ID to delete
 * @param {Request} request - Remix request with credentials
 * @returns Promise that resolves when deletion is complete
 */
export async function deleteEnvironment(id: string, request: Request): Promise<void> {
  // Build headers with authentication if request is provided
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
  
  const response = await fetch(`${API_URL}/environments/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to delete environment: ${response.statusText}`);
  }
}
