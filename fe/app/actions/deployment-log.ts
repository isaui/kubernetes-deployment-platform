
/**
 * Returns the URL for streaming build logs
 */
export function getBuildLogUrl(deploymentId: string): string {
  // Construct the build logs endpoint URL

  return `/api/v1/deployments/${deploymentId}/logs/build`;
}

/**
 * Returns the URL for streaming runtime logs
 */
export function getRuntimeLogUrl(deploymentId: string): string {
    return `/api/v1/deployments/${deploymentId}/logs/runtime`;
}