// types/service-update.ts - UPDATED untuk backend compatibility
import { Service, ServiceType, ManagedServiceType } from './service';

/**
 * Base service update DTO with common fields for all service types
 * Corresponds to BaseServiceUpdateRequest di backend
 */
export interface BaseServiceUpdate {
  name?: string;
  cpuLimit?: string;
  memoryLimit?: string;
  isStaticReplica?: boolean;
  replicas?: number;
  minReplicas?: number;
  maxReplicas?: number;
  customDomain?: string;
}

/**
 * Git service specific update fields
 * Corresponds to GitServiceUpdateRequest di backend
 */
export interface GitServiceUpdate extends BaseServiceUpdate {
  type: 'git';
  branch?: string;
  port?: number;
  buildCommand?: string;
  startCommand?: string;
  // Git services boleh update envVars
  envVars?: Record<string, string>;
}

/**
 * Managed service specific update fields  
 * Corresponds to ManagedServiceUpdateRequest di backend
 */
export interface ManagedServiceUpdate extends BaseServiceUpdate {
  type: 'managed';
  version?: string;
  storageSize?: string;
  // Managed services TIDAK boleh update envVars (auto-generated)
  envVars?: never;
}

/**
 * Combined type for service updates
 * The type discriminator determines which specific fields are available
 */
export type ServiceUpdate = GitServiceUpdate | ManagedServiceUpdate;

/**
 * Backend API format untuk service updates
 * Sesuai dengan ServiceUpdateRequest di backend DTO
 */
export interface ServiceUpdatePayload {
  type: 'git' | 'managed';
  git?: {
    name?: string;
    envVars?: Record<string, string>;
    cpuLimit?: string;
    memoryLimit?: string;
    isStaticReplica?: boolean;
    replicas?: number;
    minReplicas?: number;
    maxReplicas?: number;
    customDomain?: string;
    branch?: string;
    port?: number;
    buildCommand?: string;
    startCommand?: string;
  };
  managed?: {
    name?: string;
    cpuLimit?: string;
    memoryLimit?: string;
    isStaticReplica?: boolean;
    replicas?: number;
    minReplicas?: number;
    maxReplicas?: number;
    customDomain?: string;
    version?: string;
    storageSize?: string;
    // envVars tidak ada untuk managed services
  };
}

/**
 * Creates a service update object from a service
 * @param service The service to create an update from
 * @param fieldsToUpdate Optional array of fields to include
 */
export function createServiceUpdate(
  service: Service, 
  fieldsToUpdate?: string[]
): ServiceUpdate {
  const shouldInclude = (field: string): boolean => 
    !fieldsToUpdate || fieldsToUpdate.includes(field);

  // Build common fields
  const baseUpdate: BaseServiceUpdate = {};

  if (shouldInclude('name')) baseUpdate.name = service.name;
  if (shouldInclude('cpuLimit')) baseUpdate.cpuLimit = service.cpuLimit;
  if (shouldInclude('memoryLimit')) baseUpdate.memoryLimit = service.memoryLimit;
  if (shouldInclude('isStaticReplica')) baseUpdate.isStaticReplica = service.isStaticReplica;
  if (shouldInclude('replicas')) baseUpdate.replicas = service.replicas;
  if (shouldInclude('minReplicas')) baseUpdate.minReplicas = service.minReplicas;
  if (shouldInclude('maxReplicas')) baseUpdate.maxReplicas = service.maxReplicas;
  if (shouldInclude('customDomain')) baseUpdate.customDomain = service.customDomain || undefined;

  // Type-specific fields
  if (service.type === 'git') {
    const gitUpdate: GitServiceUpdate = {
      ...baseUpdate,
      type: 'git',
    };
    
    if (shouldInclude('envVars')) gitUpdate.envVars = service.envVars;
    if (shouldInclude('branch')) gitUpdate.branch = service.branch;
    if (shouldInclude('port')) gitUpdate.port = service.port;
    if (shouldInclude('buildCommand')) gitUpdate.buildCommand = service.buildCommand || undefined;
    if (shouldInclude('startCommand')) gitUpdate.startCommand = service.startCommand || undefined;
    
    return gitUpdate;
  } else {
    const managedUpdate: ManagedServiceUpdate = {
      ...baseUpdate,
      type: 'managed',
    };
    
    // Managed services specific fields
    if (shouldInclude('version')) managedUpdate.version = service.version;
    if (shouldInclude('storageSize')) managedUpdate.storageSize = service.storageSize;
    
    // Note: envVars tidak di-include untuk managed services karena auto-generated
    
    return managedUpdate;
  }
}

/**
 * Converts frontend ServiceUpdate to backend API format
 * @param update Frontend ServiceUpdate object
 * @returns Backend-compatible ServiceUpdatePayload
 */
export function formatUpdatePayload(update: ServiceUpdate): ServiceUpdatePayload {
  if (update.type === 'git') {
    const {
      name, envVars, cpuLimit, memoryLimit,
      isStaticReplica, replicas, minReplicas, maxReplicas,
      customDomain, branch, port, buildCommand, startCommand
    } = update;
    
    return {
      type: 'git',
      git: {
        name,
        envVars,
        cpuLimit,
        memoryLimit,
        isStaticReplica,
        replicas,
        minReplicas,
        maxReplicas,
        customDomain,
        branch,
        port,
        buildCommand,
        startCommand,
      }
    };
  } else {
    const {
      name, cpuLimit, memoryLimit,
      isStaticReplica, replicas, minReplicas, maxReplicas,
      customDomain, version, storageSize
    } = update;
    
    return {
      type: 'managed',
      managed: {
        name,
        cpuLimit,
        memoryLimit,
        isStaticReplica,
        replicas,
        minReplicas,
        maxReplicas,
        customDomain,
        version,
        storageSize,
        // Note: envVars tidak di-include untuk managed services
      }
    };
  }
}

/**
 * Validates a service update based on service type
 * @param update The service update to validate
 * @returns Array of validation errors (empty if valid)
 */
export function validateServiceUpdate(update: ServiceUpdate): string[] {
  const errors: string[] = [];
  
  if (update.type === 'managed') {
    // Managed services tidak boleh update envVars
    if ('envVars' in update && update.envVars) {
      errors.push('Environment variables are auto-generated for managed services and cannot be modified');
    }
    
    // Validate storage size format
    if (update.storageSize && !/^\d+[KMGT]?i?$/.test(update.storageSize)) {
      errors.push('Invalid storage size format. Use format like 1Gi, 500Mi, etc.');
    }
  }
  
  // Common validations
  if (update.cpuLimit && !/^\d+m?$/.test(update.cpuLimit)) {
    errors.push('Invalid CPU limit format. Use format like 500m, 1000m, etc.');
  }
  
  if (update.memoryLimit && !/^\d+[KMGT]?i$/.test(update.memoryLimit)) {
    errors.push('Invalid memory limit format. Use format like 512Mi, 1Gi, etc.');
  }
  
  return errors;
}