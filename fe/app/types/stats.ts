// Types for Kubernetes statistics responses

export interface PodStats {
  name: string;
  namespace: string;
  status: string;
  cpu: string | {
    usage: string;
    percentage: number;
    request?: string;
    limit?: string;
  };
  memory: string | {
    usage: string;
    percentage: number;
    request?: string;
    limit?: string;
  };
  created: string;
}

// Type for node condition (Ready, DiskPressure, etc)
export interface NodeCondition {
  status: string;
  reason?: string;
  message?: string;
  lastTransitionTime?: string;
}

// Type for node conditions object
export interface NodeConditions {
  [key: string]: NodeCondition;
}

// Type for resource (CPU, Memory, Storage)
export interface NodeResource {
  capacity: string;
  allocatable: string;
  usage: string;
  percentage: number;
}

export interface NodeStats {
  name: string;
  status: string;
  conditions?: NodeConditions;
  roles: string[];
  created: string;
  kubeletVersion?: string;
  osImage?: string;
  
  // Enhanced resource information
  cpu: NodeResource;
  memory: NodeResource;
  storage?: NodeResource;
  
  // Original Kubernetes node object data
  rawData?: any;
}

export interface DeploymentStats {
  name: string;
  namespace: string;
  replicas: string;
  strategy: string;
  created: string;
}

export interface ServicePort {
  name?: string;
  protocol: string;
  port: number;
  targetPort: number;
  nodePort?: number;
}

export interface ServiceStats {
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  ports: ServicePort[];
  created: string;
}

export interface IngressStats {
  name: string;
  namespace: string;
  hosts: string[];
  tls: boolean;
  class?: string;
  created: string;
  rules: number;
  paths: number;
}

export interface CertificateStats {
  name: string;
  namespace: string;
  issuer: string;
  secretName: string;
  domains: string[];
  status: string;
  isExpired: boolean;
  daysUntilExpiry: number;
  created: string;
}

export interface ClusterVersion {
  gitVersion: string;
  platform: string;
  goVersion: string;
  buildDate: string;
}

export interface ClusterStats {
  nodeCount: number;
  namespaceCount: number;
  podCount: number;
}

export interface ClusterInfo {
  version: ClusterVersion;
  stats: ClusterStats;
}

export interface PVCStats {
  name: string;
  namespace: string;
  status: string;
  storageCapacity: string;
  storageClassName?: string;
  accessModes: string[];
  volumeName?: string;
  created: string;
  phase: string;
  annotations?: string[];
}
