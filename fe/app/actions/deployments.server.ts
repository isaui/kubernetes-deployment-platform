import { Deployment } from '~/types/deployment';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Fetch deployment history for a service
 * 
 * @param {string} serviceId - Service ID to fetch deployments for
 * @param {Request} request - Remix request with credentials
 * @returns List of deployments
 */
export async function getServiceDeployments(serviceId: string, request: Request): Promise<Deployment[]> {
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
  
  const response = await fetch(`${API_URL}/services/${serviceId}/deployments`, {
    method: 'GET',
    credentials: 'include',
    headers
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch deployments: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data.deployments || [];
}

/**
 * Trigger a manual deployment for a service
 * 
 * @param {string} serviceId - Service ID to deploy
 * @param {object} deployOptions - Deployment options (commitId, etc.)
 * @param {Request} request - Remix request with credentials
 * @returns Deployment information
 */
export async function createDeployment(
  serviceId: string, 
  deployOptions: {
    commitId?: string;
    commitMessage?: string;
    callbackUrl?: string;
    apiKey?: string;
  }, 
  request: Request
): Promise<Deployment> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  };
  
  // Forward authorization cookie if request provided
  if (request?.headers) {
    const cookie = request.headers.get('Cookie');
    if (cookie) {
      headers['Cookie'] = cookie;
    }
  }
  
  const response = await fetch(`${API_URL}/deployments/git`, {
    method: 'POST',
    credentials: 'include',
    headers,
    body: JSON.stringify({
      serviceId,
      apiKey: deployOptions.apiKey || '',
      commitId: deployOptions.commitId || '',
      commitMessage: deployOptions.commitMessage || '',
    })
  });

  if (!response.ok) {
    throw new Error(`Failed to create deployment: ${response.statusText}`);
  }

  const data = await response.json();
  return data.data;
}
