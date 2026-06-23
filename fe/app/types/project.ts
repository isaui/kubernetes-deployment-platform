export interface Environment {
  id: string;
  name: string;
  description: string;
  projectId: string;
  createdAt: string;
  updatedAt: string;
  services?: Service[];
}

export interface Project {
  id: string;
  name: string;
  description: string;
  userId: string;
  createdAt: string;
  updatedAt: string;
  user?: User;
  environments?: Environment[];
  services?: Service[];
}

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
}

export interface Service {
  id: string;
  name: string;
  type: string;
  gitRepo: string;
  buildCommand: string | null;
  startCommand: string | null;
  envVars: { key: string; value: string }[];
  status: string;
  domain: string | null;
  projectID: string;
  environmentId: string;
  createdAt: string;
  updatedAt: string;
  isStaticReplica: boolean;
  minReplicas: number;
  maxReplicas: number;
  replicas: number;
  deployments?: Deployment[];
}

export interface Deployment {
  id: string;
  serviceID: string;
  status: string;
  buildLog: string | null;
  deploymentLog: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectFilter {
  search?: string;
  sortBy?: string;
  sortOrder?: string;
  page: number;
  pageSize: number;
}

export interface ProjectListResponse {
  projects: Project[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ProjectEnvironmentItem {
  id: string;
  name: string;
  description: string;
  servicesCount: number;
  createdAt: string;
}

export interface ProjectServiceStatsItem {
  id: string;
  name: string;
  type: string;
  status: string;
  environmentId: string;
  environmentName: string;
  deployments: number;
  successRate: number;
  replicas: number;
  isAutoScaling: boolean;
}

export interface ProjectStatsResponse {
  project: {
    id: string;
    name: string;
    description: string;
    createdAt: string;
  };
  environments: {
    total: number;
    environments: ProjectEnvironmentItem[];
  };
  services: {
    total: number;
    byType: Record<string, number>;
    byStatus: Record<string, number>;
    servicesList: ProjectServiceStatsItem[];
  };
  deployments: {
    total: number;
    successful: number;
    failed: number;
    inProgress: number;
    successRate: number;
  };
}
