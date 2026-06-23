import { Deployment } from '~/types/deployment';

/**
 * Fetch deployment history for a service
 * 
 * @param {string} serviceId - Service ID to fetch deployments for
 * @returns List of deployments
 */
export async function getServiceDeployments(serviceId: string): Promise<Deployment[]> {
  const response = await fetch(`/api/v1/services/${serviceId}/deployments`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }
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
): Promise<Deployment> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  };
  
  const response = await fetch(`/api/v1/deployments/git`, {
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
