import { getEnv } from '~/utils/env.server';
import type { 
  PodStats, 
  NodeStats, 
  DeploymentStats, 
  ServiceStats, 
  IngressStats, 
  CertificateStats,
  ClusterInfo,
  PVCStats 
} from '~/types/stats';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Handle API error responses
 * 
 * @param {Response} response - The fetch response
 * @param {string} resourceType - The type of resource being fetched
 */
async function handleApiError(response: Response, resourceType: string) {
  if (response.status === 401) {
    throw new Response('Unauthorized', { status: 401 });
  }
  if (response.status === 403) {
    throw new Response('Forbidden - Admin access required', { status: 403 });
  }
  const error = await response.text();
  throw new Response(`Error fetching ${resourceType}: ${error}`, { status: response.status });
}

/**
 * Fetch PVC (Persistent Volume Claim) statistics
 * 
 * @param {Request} request - Remix request with credentials
 * @param {string} namespace - Optional namespace to filter PVCs
 * @returns PVC statistics data
 */
export async function getPVCStats(request: Request, namespace?: string): Promise<{ pvcs: PVCStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/pvc`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      await handleApiError(response, 'PVC stats');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch PVC statistics', { status: 500 });
  }
}

/**
 * Fetch pod statistics
 * 
 * @param {string} namespace - Optional namespace to filter pods
 * @param {Request} request - Remix request with credentials
 * @returns Pod statistics data
 */
export async function getPodStats(request: Request, namespace?: string): Promise<{ pods: PodStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/pods`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new Response('Unauthorized', { status: 401 });
      }
      if (response.status === 403) {
        throw new Response('Forbidden - Admin access required', { status: 403 });
      }
      const error = await response.text();
      throw new Response(`Error fetching pod stats: ${error}`, { status: response.status });
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch pod statistics', { status: 500 });
  }
}

/**
 * Fetch node statistics
 * 
 * @param {Request} request - Remix request with credentials
 * @returns Node statistics data that's already processed by the backend
 */
export async function getNodeStats(request: Request): Promise<{ nodes: NodeStats[] }> {
  try {
    const response = await fetch(`${API_URL}/admin/stats/nodes`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new Response('Unauthorized', { status: 401 });
      }
      if (response.status === 403) {
        throw new Response('Forbidden - Admin access required', { status: 403 });
      }
      const error = await response.text();
      throw new Response(`Error fetching node stats: ${error}`, { status: response.status });
    }
    
    const data = await response.json();
    // PENTING: Biarkan nilai usage dari backend tetap ada
    // Backend sudah mengirim nilai actual usage (dari metrics api)
    if (data.nodes && Array.isArray(data.nodes)) {
      data.nodes.forEach((node: any) => {
        // Hanya set usage jika tidak ada dari backend
        if (node.cpu && !node.cpu.usage) {
          node.cpu.usage = node.cpu.usage || "0";
        }
        if (node.memory && !node.memory.usage) {
          node.memory.usage = node.memory.usage || "0";
        }
        if (node.storage && !node.storage.usage) {
          node.storage.usage = node.storage.usage || "0";
        }
      });
    }    
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch node statistics', { status: 500 });
  }
}

/**
 * Fetch deployment statistics
 * 
 * @param {string} namespace - Optional namespace to filter deployments
 * @param {Request} request - Remix request with credentials
 * @returns Deployment statistics data
 */
export async function getDeploymentStats(request: Request, namespace?: string): Promise<{ deployments: DeploymentStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/deployments`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new Response('Unauthorized', { status: 401 });
      }
      if (response.status === 403) {
        throw new Response('Forbidden - Admin access required', { status: 403 });
      }
      const error = await response.text();
      throw new Response(`Error fetching deployment stats: ${error}`, { status: response.status });
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch deployment statistics', { status: 500 });
  }
}

/**
 * Fetch service statistics
 * 
 * @param {string} namespace - Optional namespace to filter services
 * @param {Request} request - Remix request with credentials
 * @returns Service statistics data
 */
export async function getServiceStats(request: Request, namespace?: string): Promise<{ services: ServiceStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/services`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      await handleApiError(response, 'service stats');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch service statistics', { status: 500 });
  }
}

/**
 * Fetch ingress statistics
 * 
 * @param {string} namespace - Optional namespace to filter ingress resources
 * @param {Request} request - Remix request with credentials
 * @returns Ingress statistics data
 */
export async function getIngressStats(request: Request, namespace?: string): Promise<{ ingresses: IngressStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/ingress`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      await handleApiError(response, 'ingress stats');
    }

    const data = await response.json();
    
    // Extract hosts from TLS data if main hosts property is empty
    if (data.ingresses && Array.isArray(data.ingresses)) {
      data.ingresses = data.ingresses.map((ingress:any) => {
        // If hosts is empty or undefined but tls has hosts, use those
        if ((!ingress.hosts || ingress.hosts.length === 0) && 
            ingress.tls && Array.isArray(ingress.tls) && 
            ingress.tls.length > 0 && 
            ingress.tls[0].hosts) {
          return {
            ...ingress,
            hosts: ingress.tls[0].hosts
          };
        }
        return ingress;
      });
    }
    
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch ingress statistics', { status: 500 });
  }
}

/**
 * Fetch certificate statistics
 * 
 * @param {string} namespace - Optional namespace to filter certificates
 * @param {Request} request - Remix request with credentials
 * @returns Certificate statistics data
 */
export async function getCertificateStats(request: Request, namespace?: string): Promise<{ certificates: CertificateStats[] }> {
  try {
    const url = new URL(`${API_URL}/admin/stats/certificates`);
    if (namespace) {
      url.searchParams.append('namespace', namespace);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      await handleApiError(response, 'certificate stats');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch certificate statistics', { status: 500 });
  }
}

/**
 * Fetch cluster information
 * 
 * @param {Request} request - Remix request with credentials
 * @returns Cluster information data
 */
export async function getClusterInfo(request: Request): Promise<ClusterInfo> {
  try {
    const response = await fetch(`${API_URL}/admin/cluster/info`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Cookie': request.headers.get('Cookie') || ''
      }
    });

    if (!response.ok) {
      await handleApiError(response, 'cluster info');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to fetch cluster information', { status: 500 });
  }
}
