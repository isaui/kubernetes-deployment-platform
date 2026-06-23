/**
 * Represents a service deployment
 */
export interface Deployment {
  id: string;
  serviceId: string;
  status: 'building' | 'success' | 'failed';
  commitSha?: string;
  commitMessage?: string;
  image?: string;
  version?: string;
  createdAt: string;
}
