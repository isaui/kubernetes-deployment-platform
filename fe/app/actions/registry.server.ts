import type { Registry, RegistryFilter, CreateRegistryRequest, UpdateRegistryRequest, RegistryListResponse, RegistryDetailResponse } from '~/types/registry';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Helper function to get headers with authentication
 */
function getHeaders(request: Request): HeadersInit {
  const headers: HeadersInit = {
    'Accept': 'application/json',
    'Content-Type': 'application/json'
  };
  
  if (request?.headers) {
    const cookie = request.headers.get('Cookie');
    if (cookie) {
      headers['Cookie'] = cookie;
    }
  }
  
  return headers;
}

/**
 * Fetch list of registries with pagination and filtering
 * 
 * @param {RegistryFilter} filter - Filter and pagination options
 * @param {Request} request - Remix request with credentials
 * @returns List of registries and pagination info
 */
export async function getRegistries(filter: RegistryFilter = {}, request: Request): Promise<RegistryListResponse> {
  const queryParams = new URLSearchParams();
  
  if (filter.page) queryParams.append('page', filter.page.toString());
  if (filter.pageSize) queryParams.append('pageSize', filter.pageSize.toString());
  if (filter.search) queryParams.append('search', filter.search);
  if (filter.sortBy) queryParams.append('sortBy', filter.sortBy);
  if (filter.sortOrder) queryParams.append('sortOrder', filter.sortOrder);
  if (filter.onlyActive !== undefined) queryParams.append('onlyActive', filter.onlyActive.toString());
  
  const url = `${API_URL}/registries?${queryParams.toString()}`;
  
  const response = await fetch(url, {
    method: 'GET',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request)
  });
  
  if (!response.ok) {
    throw new Error(`Failed to fetch registries: ${response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Fetch a single registry by ID
 * 
 * @param {string} id - Registry ID to fetch
 * @param {Request} request - Remix request with credentials
 * @returns Registry details
 */
export async function getRegistry(id: string, request: Request): Promise<Registry> {
  const response = await fetch(`${API_URL}/registries/${id}`, {
    method: 'GET',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request)
  });
  
  if (!response.ok) {
    throw new Error(`Failed to fetch registry: ${response.statusText}`);
  }
  
  return await response.json();
}

export async function getRegistryDetails(id: string, request: Request): Promise<RegistryDetailResponse> {
  const response = await fetch(`${API_URL}/registries/${id}/details`, {
    method: 'GET',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request)
  });
  
  if (!response.ok) {
    throw new Error(`Failed to fetch registry: ${response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Create a new registry
 * 
 * @param {CreateRegistryRequest} registry - Registry data to create
 * @param {Request} request - Remix request with credentials
 * @returns Newly created registry
 */
export async function createRegistry(registry: CreateRegistryRequest, request: Request): Promise<Registry> {
  const response = await fetch(`${API_URL}/registries`, {
    method: 'POST',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request),
    body: JSON.stringify(registry)
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to create registry: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Update an existing registry
 * 
 * @param {string} id - Registry ID to update
 * @param {UpdateRegistryRequest} registry - Updated registry data
 * @param {Request} request - Remix request with credentials
 * @returns Updated registry
 */
export async function updateRegistry(id: string, registry: UpdateRegistryRequest, request: Request): Promise<Registry> {
  const response = await fetch(`${API_URL}/registries/${id}`, {
    method: 'PUT',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request),
    body: JSON.stringify(registry)
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to update registry: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Delete a registry
 * 
 * @param {string} id - Registry ID to delete
 * @param {Request} request - Remix request with credentials
 */
export async function deleteRegistry(id: string, request: Request): Promise<void> {
  const response = await fetch(`${API_URL}/registries/${id}`, {
    method: 'DELETE',
    credentials: 'include', // Include cookies for authentication
    headers: getHeaders(request)
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to delete registry: ${errorData.error || response.statusText}`);
  }
}

/**
 * Get registry build logs URL for Server-Sent Events (SSE) streaming
 * 
 * @param {string} id - Registry ID to get logs for
 * @returns Full URL to the SSE streaming logs endpoint
 */
export function getRegistryBuildLogsUrl(id: string): string {
  return `${API_URL}/registries/${id}/logs/stream`;
}
