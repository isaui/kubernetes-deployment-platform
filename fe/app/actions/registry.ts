import type { Registry, RegistryFilter, CreateRegistryRequest, UpdateRegistryRequest, RegistryListResponse } from '~/types/registry';

/**
 * Client-side function to fetch registries with filters
 * 
 * @param {RegistryFilter} filter - Filter and pagination options
 * @returns List of registries and pagination info
 */
export async function fetchRegistries(filter: RegistryFilter = {}): Promise<RegistryListResponse> {
  const queryParams = new URLSearchParams();
  
  if (filter.page) queryParams.append('page', filter.page.toString());
  if (filter.pageSize) queryParams.append('pageSize', filter.pageSize.toString());
  if (filter.search) queryParams.append('search', filter.search);
  if (filter.sortBy) queryParams.append('sortBy', filter.sortBy);
  if (filter.sortOrder) queryParams.append('sortOrder', filter.sortOrder);
  if (filter.onlyActive !== undefined) queryParams.append('onlyActive', filter.onlyActive.toString());
  
  const response = await fetch(`/api/v1/registries?${queryParams.toString()}`, {
    method: 'GET',
    credentials: 'include', // Include cookies for authentication
    headers: {
      'Accept': 'application/json'
    }
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to fetch registries: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Client-side function to fetch a single registry
 * 
 * @param {string} id - Registry ID to fetch
 * @returns Registry details
 */
export async function fetchRegistry(id: string): Promise<Registry> {
  const response = await fetch(`/api/v1/registries/${id}`, {
    method: 'GET',
    credentials: 'include', // Include cookies for authentication
    headers: {
      'Accept': 'application/json'
    }
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to fetch registry: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Client-side function to create a new registry
 * 
 * @param {CreateRegistryRequest} registry - Registry data to create
 * @returns Newly created registry
 */
export async function createRegistry(registry: CreateRegistryRequest): Promise<Registry> {
  const response = await fetch(`/api/v1/registries`, {
    method: 'POST',
    credentials: 'include', // Include cookies for authentication
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    },
    body: JSON.stringify(registry)
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to create registry: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Client-side function to update an existing registry
 * 
 * @param {string} id - Registry ID to update
 * @param {UpdateRegistryRequest} registry - Updated registry data
 * @returns Updated registry
 */
export async function updateRegistry(id: string, registry: UpdateRegistryRequest): Promise<Registry> {
  const response = await fetch(`/api/v1/registries/${id}`, {
    method: 'PUT',
    credentials: 'include', // Include cookies for authentication
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    },
    body: JSON.stringify(registry)
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to update registry: ${errorData.error || response.statusText}`);
  }
  
  return await response.json();
}

/**
 * Client-side function to delete a registry
 * 
 * @param {string} id - Registry ID to delete
 */
export async function deleteRegistry(id: string): Promise<void> {
  const response = await fetch(`/api/v1/registries/${id}`, {
    method: 'DELETE',
    credentials: 'include', // Include cookies for authentication
  });
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(`Failed to delete registry: ${errorData.error || response.statusText}`);
  }
}

/**
 * Client-side function to setup SSE connection for build logs using EventSource
 * 
 * @param {string} id - Registry ID to stream logs for
 * @param {function} onMessage - Callback function for log messages
 * @param {function} onError - Callback function for errors
 * @param {function} onComplete - Callback function when streaming is complete
 * @returns EventSource instance and a close function
 */
export function setupBuildLogsStream(
  id: string, 
  onMessage: (message: string) => void,
  onError?: (event: Event) => void,
  onComplete?: () => void
): { eventSource: EventSource; close: () => void } {
  // Create EventSource connection to stream logs
  const url = `/api/v1/registries/${id}/logs/stream`;
  const eventSource = new EventSource(url, { withCredentials: true });
  
  // Handle incoming messages
  eventSource.onmessage = (event) => {
    onMessage(event.data);
  };
  
  // Handle errors
  if (onError) {
    eventSource.onerror = onError;
  }
  
  // Handle stream completion (custom event from server)
  eventSource.addEventListener('complete', () => {
    if (onComplete) {
      onComplete();
    }
    eventSource.close();
  });
  
  // Return eventSource and close function
  return {
    eventSource,
    close: () => eventSource.close()
  };
}
