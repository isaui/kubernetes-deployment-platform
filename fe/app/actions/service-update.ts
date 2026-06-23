// api/service-update.ts - SIMPLIFIED & backend compatible
import { ServiceUpdate, formatUpdatePayload, validateServiceUpdate } from "~/types/service-update";

/**
 * Updates a service's configuration via API
 * 
 * @param serviceId The ID of the service to update
 * @param updateData The update data for the service
 * @returns The updated service data
 */
export async function updateService(serviceId: string, updateData: ServiceUpdate) {
  // Validate the update data first
  const validationErrors = validateServiceUpdate(updateData);
  if (validationErrors.length > 0) {
    throw new Error(`Validation failed: ${validationErrors.join(', ')}`);
  }
  
  // Convert frontend ServiceUpdate to backend format
  const payload = formatUpdatePayload(updateData);
  
  console.log('Sending service update payload:', payload); // Debug log
  
  const response = await fetch(`/api/v1/services/${serviceId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
    credentials: 'include'
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.error || `Failed to update service: ${response.statusText}`);
  }

  const result = await response.json();
  return result.data;
}

/**
 * Helper function to create a partial update for a service
 * Only includes changed fields to minimize payload size
 * 
 * @param original Original service data
 * @param updated Updated service data
 * @returns ServiceUpdate with only changed fields
 */
export function createPartialUpdate(
  original: any, 
  updated: any
): ServiceUpdate | null {
  const changes: any = { type: original.type };
  let hasChanges = false;
  
  // Common fields that can be updated
  const commonFields = [
    'name', 'cpuLimit', 'memoryLimit', 
    'isStaticReplica', 'replicas', 'minReplicas', 'maxReplicas',
    'customDomain'
  ];
  
  // Check common fields for changes
  commonFields.forEach(field => {
    if (updated[field] !== undefined && updated[field] !== original[field]) {
      changes[field] = updated[field];
      hasChanges = true;
    }
  });
  
  // Type-specific fields
  if (original.type === 'git') {
    const gitFields = ['branch', 'port', 'buildCommand', 'startCommand', 'envVars'];
    
    gitFields.forEach(field => {
      if (updated[field] !== undefined && updated[field] !== original[field]) {
        changes[field] = updated[field];
        hasChanges = true;
      }
    });
  } else if (original.type === 'managed') {
    const managedFields = ['version', 'storageSize'];
    
    managedFields.forEach(field => {
      if (updated[field] !== undefined && updated[field] !== original[field]) {
        changes[field] = updated[field];
        hasChanges = true;
      }
    });
  }
  console.log(hasChanges)
  console.log('Partial update:', changes);
  
  return hasChanges ? changes as ServiceUpdate : null;
}

/**
 * Validates service update payload before sending to backend
 * 
 * @param update ServiceUpdate object to validate
 * @returns True if valid, throws error if invalid
 */
export function validateServiceUpdatePayload(update: ServiceUpdate): boolean {
  
  // Check required type field
  if (!update.type || !['git', 'managed'].includes(update.type)) {
    throw new Error('Service type is required and must be either "git" or "managed"');
  }
  
  // Type-specific validation
  if (update.type === 'managed') {
    // Managed services cannot update envVars
    if ('envVars' in update && update.envVars) {
      throw new Error('Environment variables cannot be updated for managed services');
    }
    
    // Validate version format if provided
    if (update.version && !update.version.trim()) {
      throw new Error('Version cannot be empty');
    }
    
    // Validate storage size format if provided
    if (update.storageSize && !/^\d+[KMGT]?i?$/.test(update.storageSize)) {
      throw new Error('Invalid storage size format. Use format like 1Gi, 500Mi, etc.');
    }
  }
  
  // Resource validation
  if (update.cpuLimit && !/^\d+m?$/.test(update.cpuLimit)) {
    throw new Error('Invalid CPU limit format. Use format like 500m, 1000m, etc.');
  }
  
  if (update.memoryLimit && !/^\d+[KMGT]?i$/.test(update.memoryLimit)) {
    throw new Error('Invalid memory limit format. Use format like 512Mi, 1Gi, etc.');
  }
  
  // Replica validation
  if (update.replicas !== undefined && update.replicas < 1) {
    throw new Error('Replicas must be at least 1');
  }
  
  if (update.minReplicas !== undefined && update.minReplicas < 1) {
    throw new Error('Minimum replicas must be at least 1');
  }
  
  if (update.maxReplicas !== undefined && update.maxReplicas < 1) {
    throw new Error('Maximum replicas must be at least 1');
  }
  
  if (update.minReplicas !== undefined && update.maxReplicas !== undefined && 
      update.minReplicas > update.maxReplicas) {
    throw new Error('Minimum replicas cannot be greater than maximum replicas');
  }
  
  return true;
}