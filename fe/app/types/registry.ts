// Types for Registry API
export type RegistryStatus = 'pending' | 'building' | 'ready' | 'failed';

export interface Registry {
  id: string;
  name: string;
  url: string;
  isDefault: boolean;
  isActive: boolean;
  status: RegistryStatus;
  createdAt: string;
  updatedAt: string;
}

export interface RegistryFilter {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: string;
  sortOrder?: string;
  onlyActive?: boolean;
}

export interface CreateRegistryRequest {
  name: string;
  isDefault?: boolean;
}

export interface UpdateRegistryRequest {
  name?: string;
  isDefault?: boolean;
  isActive?: boolean;
}

export interface RegistryListResponse {
  registries: Registry[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface RegistryCredentials {
  url?: string;
}

export interface RegistryImageInfo {
  name: string;
  tags: string[];
  size: number;
  created: string;
  digest: string;
}

export interface RegistryDetailResponse {
  registry: Registry;
  credentials?: RegistryCredentials;
  images: RegistryImageInfo[];
  imagesCount: number;
  size: number;
  isHealthy: boolean;
  kubeStatus: string;
  lastSynced: string;
}