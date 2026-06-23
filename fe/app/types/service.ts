// types/service.ts - UPDATED untuk managed services
import { Environment } from "./project";

export type ServiceType = 'git' | 'managed';

// Supported managed service types (sync dengan backend)
export type ManagedServiceType = 
  | 'postgresql' 
  | 'mysql' 
  | 'redis' 
  | 'mongodb' 
  | 'minio' 
  | 'rabbitmq';

export interface Service {
  // Common fields for all service types
  id: string;
  name: string;
  type: ServiceType;
  projectId: string;
  environmentId: string;
  
  // Git-specific fields (only relevant when type is 'git')
  repoUrl?: string;
  branch?: string;
  isPublic?: boolean;
  gitUsername?: string;
  // gitToken is never returned by the backend (write-only)
  port?: number;
  buildCommand?: string | null;
  startCommand?: string | null;
  
  // Managed service fields (only relevant when type is 'managed')
  managedType?: ManagedServiceType;  // postgresql, redis, minio, etc.
  version?: string;                  // 14, 6.0, latest, etc.
  storageSize?: string;             // 1Gi, 10Gi, etc.
  
  // Resources & Scaling
  cpuLimit: string;
  memoryLimit: string;
  isStaticReplica: boolean;
  replicas: number;
  minReplicas: number;
  maxReplicas: number;
  
  // Domain
  domain: string | null;
  customDomain: string | null;
  
  // Status
  status: string; // inactive, building, running, failed
  
  // API Key for webhooks (hanya untuk git services)
  apiKey?: string;
  
  // Environment Variables (read-only untuk managed services)
  envVars?: Record<string, string>;
  
  createdAt: string;
  updatedAt: string;
  
  // Relations
  environment?: Environment;
}

export interface ServiceListResponse {
  data: {
    services: Service[];
  };
}

export interface ServiceResponse {
  data: Service;
}

// Service creation payload for API
export interface CreateServicePayload {
  // Common fields
  name: string;
  type: ServiceType;
  projectId: string;
  environmentId: string;
  
  // Optional common fields
  cpuLimit?: string;
  memoryLimit?: string;
  isStaticReplica?: boolean;
  replicas?: number;
  minReplicas?: number;
  maxReplicas?: number;
  customDomain?: string;
  
  // Git-specific fields (required when type is 'git')
  repoUrl?: string;
  branch?: string;
  isPublic?: boolean;
  gitUsername?: string;
  gitToken?: string;
  port?: number;
  buildCommand?: string;
  startCommand?: string;
  envVars?: Record<string, string>;
  
  // Managed service fields (required when type is 'managed')
  managedType?: ManagedServiceType;
  version?: string;
  storageSize?: string;
}

// Helper functions
export function isGitService(service: Service): service is Service & { 
  repoUrl: string; 
  branch: string; 
  apiKey: string; 
} {
  return service.type === 'git';
}

export function isManagedService(service: Service): service is Service & { 
  managedType: ManagedServiceType; 
} {
  return service.type === 'managed';
}

// Managed service configurations (sync dengan backend)
export const MANAGED_SERVICE_CONFIGS = {
  postgresql: {
    name: 'PostgreSQL',
    defaultVersion: '15',
    defaultPort: 5432,
    description: 'Relational database',
    requiresStorage: true,
  },
  mysql: {
    name: 'MySQL',
    defaultVersion: '8.0',
    defaultPort: 3306,
    description: 'Relational database',
    requiresStorage: true,
  },
  redis: {
    name: 'Redis',
    defaultVersion: '7',
    defaultPort: 6379,
    description: 'In-memory data store',
    requiresStorage: true,
  },
  mongodb: {
    name: 'MongoDB',
    defaultVersion: '7.0',
    defaultPort: 27017,
    description: 'Document database',
    requiresStorage: true,
  },
  minio: {
    name: 'MinIO',
    defaultVersion: 'latest',
    defaultPort: 9000,
    description: 'Object storage',
    requiresStorage: true,
  },
  rabbitmq: {
    name: 'RabbitMQ',
    defaultVersion: '3.12',
    defaultPort: 5672,
    description: 'Message broker',
    requiresStorage: true,
  },
} as const;

export function getManagedServiceConfig(type: ManagedServiceType) {
  return MANAGED_SERVICE_CONFIGS[type];
}