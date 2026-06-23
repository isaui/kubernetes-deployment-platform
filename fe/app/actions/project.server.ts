import type { Project, ProjectFilter, ProjectListResponse, ProjectStatsResponse, Environment } from '~/types/project';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Fetch list of projects with filtering and pagination
 * 
 * @param {ProjectFilter} filter - Filter options for projects
 * @param {Request} request - Remix request with credentials
 * @returns Project list response with pagination
 */
export async function getProjects(filter: ProjectFilter, request: Request): Promise<ProjectListResponse> {
  const params = new URLSearchParams();
  
  if (filter.search) params.append('search', filter.search);
  if (filter.sortBy) params.append('sortBy', filter.sortBy);
  if (filter.sortOrder) params.append('sortOrder', filter.sortOrder);
  
  params.append('page', filter.page.toString());
  params.append('pageSize', filter.pageSize.toString());

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
  
  const response = await fetch(`${API_URL}/projects?${params.toString()}`, {
    method: 'GET',
    credentials: 'include',
    headers
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch projects: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Fetch a single project by ID
 * 
 * @param {string} id - Project ID to fetch
 * @param {Request} request - Remix request with credentials
 * @returns Project details
 */
export async function getProject(id: string, request: Request): Promise<Project> {
  const response = await fetch(`${API_URL}/projects/${id}`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
      'Cookie': request.headers.get('Cookie') || ''
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch project: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Get project statistics for dashboard
 * 
 * @param {string} id - Project ID to fetch stats for
 * @param {Request} request - Remix request with credentials
 * @returns Project statistics
 */
export async function getProjectStats(id: string, request: Request): Promise<ProjectStatsResponse> {
  const response = await fetch(`${API_URL}/projects/${id}/stats`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
      'Cookie': request.headers.get('Cookie') || ''
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch project stats: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Create a new project
 * 
 * @param {Partial<Project>} project - Project data to create
 * @param {Request} request - Remix request with credentials
 * @returns Newly created project
 */
export async function createProject(project: Partial<Project>, request: Request): Promise<Project> {
  const response = await fetch(`${API_URL}/projects`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Cookie': request.headers.get('Cookie') || ''
    },
    body: JSON.stringify(project)
  });

  if (!response.ok) {
    throw new Error(`Failed to create project: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Update an existing project
 * 
 * @param {string} id - Project ID to update
 * @param {Partial<Project>} project - Updated project data
 * @param {Request} request - Remix request with credentials
 * @returns Updated project
 */
export async function updateProject(id: string, project: Partial<Project>, request: Request): Promise<Project> {
  const response = await fetch(`${API_URL}/projects/${id}`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Cookie': request.headers.get('Cookie') || ''
    },
    body: JSON.stringify(project)
  });

  if (!response.ok) {
    throw new Error(`Failed to update project: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * Delete a project
 * 
 * @param {string} id - Project ID to delete
 * @param {Request} request - Remix request with credentials
 * @returns Promise that resolves when deletion is complete
 */
export async function deleteProject(id: string, request: Request): Promise<void> {
  const response = await fetch(`${API_URL}/projects/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
      'Cookie': request.headers.get('Cookie') || ''
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to delete project: ${response.statusText}`);
  }
}
