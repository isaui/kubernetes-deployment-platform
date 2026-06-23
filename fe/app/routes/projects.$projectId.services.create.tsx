import { useState, useEffect, useRef } from 'react';
import { json, redirect, type ActionFunctionArgs, type LoaderFunctionArgs } from '@remix-run/node';
import { Form, useLoaderData, useActionData, useNavigation, Link } from '@remix-run/react';
import { createService } from '~/actions/service.server';
import { getEnvironments } from '~/actions/environment.server';
import { ArrowLeft, Loader2, Server, GitBranch, Database, Globe, Settings, AlertCircle, Plus, Code, Package, Lock, KeyRound } from 'lucide-react';
import type { ServiceType, CreateServicePayload, ManagedServiceType } from '~/types/service';
import { MANAGED_SERVICE_CONFIGS, getManagedServiceConfig } from '~/types/service';

export async function loader({ params, request }: LoaderFunctionArgs) {
  const { projectId } = params;
  
  if (!projectId) {
    throw new Response('Project ID is required', { status: 400 });
  }
  
  try {
    // Load environments for dropdown
    const environments = await getEnvironments(projectId, request);
    
    return json({ 
      projectId, 
      environments 
    });
  } catch (error) {
    console.error('Error loading environments:', error);
    throw new Response('Failed to load environments', { status: 500 });
  }
}

export async function action({ request, params }: ActionFunctionArgs) {
  const { projectId } = params;
  
  if (!projectId) {
    return json({ error: 'Project ID is required' }, { status: 400 });
  }
  
  const formData = await request.formData();
  
  try {
    // Get common fields
    const type = formData.get('type') as ServiceType;
    const serviceData: CreateServicePayload = {
      name: formData.get('name') as string,
      type,
      projectId: projectId,
      environmentId: formData.get('environmentId') as string,
      // Default resource values
      cpuLimit: '500m',
      memoryLimit: '512Mi',
      isStaticReplica: true,
      replicas: 1
    };
    
    // Add type-specific fields
    if (type === 'git') {
      // Git-specific fields
      serviceData.repoUrl = formData.get('repoUrl') as string;
      serviceData.branch = formData.get('branch') as string || 'main';
      serviceData.port = Number(formData.get('port')) || 3000;

      // Visibility & credentials (checkbox sends 'on' only when checked)
      const isPublic = formData.get('isPublic') === 'on';
      serviceData.isPublic = isPublic;
      const gitUsername = (formData.get('gitUsername') as string || '').trim();
      const gitToken = (formData.get('gitToken') as string || '').trim();
      if (gitUsername) serviceData.gitUsername = gitUsername;
      if (gitToken) serviceData.gitToken = gitToken;

      // Validate Git-specific required fields
      if (!serviceData.name || !serviceData.environmentId || !serviceData.repoUrl) {
        return json({ error: 'Please fill in all required fields for Git service' }, { status: 400 });
      }

      // Only HTTPS URLs are supported (PAT authentication is HTTPS-only)
      if (!/^https:\/\//i.test(serviceData.repoUrl)) {
        return json({ error: 'Repository URL must be an HTTPS URL (e.g. https://github.com/owner/repo.git)' }, { status: 400 });
      }

      // Private repositories require a token
      if (!isPublic && !gitToken) {
        return json({ error: 'A personal access token is required for private repositories' }, { status: 400 });
      }
    } else if (type === 'managed') {
      // Managed service fields
      const managedType = formData.get('managedType') as ManagedServiceType;
      serviceData.managedType = managedType;
      
      // Get configured defaults for this managed service type
      const serviceConfig = getManagedServiceConfig(managedType);
      
      serviceData.version = (formData.get('version') as string) || serviceConfig?.defaultVersion || 'latest';
      serviceData.storageSize = (formData.get('storageSize') as string) || '1Gi';
      
      // Validate managed service required fields
      if (!serviceData.name || !serviceData.environmentId || !serviceData.managedType) {
        return json({ error: 'Please fill in all required fields for Managed service' }, { status: 400 });
      }
      
      // Validate storage size
      if (serviceData.storageSize) {
        // Extract numeric value and unit from storageSize (e.g., "5Gi" -> 5, "Gi")
        const sizeMatch = serviceData.storageSize.match(/^([0-9]+)([A-Za-z]+)$/);
        if (!sizeMatch) {
          return json({ error: 'Invalid storage size format. Please use format like "1Gi" or "500Mi"' }, { status: 400 });
        }
        
        const value = parseInt(sizeMatch[1], 10);
        const unit = sizeMatch[2];
        
        // Convert to Gi for validation if in Mi
        let sizeInGi = value;
        if (unit.toLowerCase() === 'mi') {
          sizeInGi = value / 1024;
        } else if (unit.toLowerCase() !== 'gi') {
          return json({ error: 'Storage size must use Gi or Mi units' }, { status: 400 });
        }
        
        // Check if exceeds 8Gi limit
        if (sizeInGi > 8) {
          return json({ error: 'Storage size cannot exceed 8Gi' }, { status: 400 });
        }
      }
    } else {
      return json({ error: 'Invalid service type' }, { status: 400 });
    }
    
    // Create service
    const service = await createService(serviceData, request);
    
    // Redirect to service detail page
    return redirect(`/services/${service.id}`);
  } catch (error) {
    console.error('Error creating service:', error);
    return json({ error: (error as Error).message || 'Failed to create service' }, { status: 500 });
  }
}

export default function CreateService() {
  const { projectId, environments } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === 'submitting';
  
  const [serviceType, setServiceType] = useState<ServiceType>('git');
  const [isPublic, setIsPublic] = useState<boolean>(true);
  const [managedType, setManagedType] = useState<ManagedServiceType | ''>('');
  const [defaultVersion, setDefaultVersion] = useState<string>('latest');
  const [storageSizeError, setStorageSizeError] = useState<string>('');
  const storageSizeRef = useRef<HTMLInputElement>(null);

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-8">
          {/* Header Section */}
          <div className="space-y-4">
            <Link 
              to={`/projects/${projectId}`}
              className="inline-flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors group"
            >
              <ArrowLeft size={16} className="mr-2 transition-transform group-hover:-translate-x-1" />
              Back to Project
            </Link>
            
            <div className="text-center space-y-3">
            
              <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
                Create Service
              </h1>
              <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                Deploy your application or set up a managed service
              </p>
            </div>
          </div>

          {/* Error Message */}
          {actionData?.error && (
            <div className="max-w-2xl mx-auto">
              <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <AlertCircle className="h-4 w-4 text-red-500 mt-0.5" />
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Creation Failed</h3>
                    <p className="text-sm text-red-700 dark:text-red-300 mt-1">{actionData.error}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Form Container */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl p-8">
              <Form method="post" className="space-y-8">
                {/* Service Type Selection */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Service Type</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {/* Git Service */}
                    <label className={`relative cursor-pointer rounded-xl border-2 p-4 transition-all ${
                      serviceType === 'git' 
                        ? 'border-indigo-500 bg-indigo-50/50 dark:bg-indigo-900/20' 
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                    }`}>
                      <input
                        type="radio"
                        name="type"
                        value="git"
                        checked={serviceType === 'git'}
                        onChange={(e) => setServiceType(e.target.value as ServiceType)}
                        className="sr-only"
                      />
                      <div className="flex items-start space-x-3">
                        <div className={`p-2 rounded-lg ${
                          serviceType === 'git' 
                            ? 'bg-indigo-600 text-white' 
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                        }`}>
                          <GitBranch className="w-5 h-5" />
                        </div>
                        <div className="flex-1">
                          <h4 className="text-sm font-semibold text-gray-900 dark:text-white">Git Repository</h4>
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Deploy from Git repository</p>
                        </div>
                      </div>
                    </label>

                    {/* Managed Service */}
                    <label className={`relative cursor-pointer rounded-xl border-2 p-4 transition-all ${
                      serviceType === 'managed' 
                        ? 'border-indigo-500 bg-indigo-50/50 dark:bg-indigo-900/20' 
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                    }`}>
                      <input
                        type="radio"
                        name="type"
                        value="managed"
                        checked={serviceType === 'managed'}
                        onChange={(e) => setServiceType(e.target.value as ServiceType)}
                        className="sr-only"
                      />
                      <div className="flex items-start space-x-3">
                        <div className={`p-2 rounded-lg ${
                          serviceType === 'managed' 
                            ? 'bg-indigo-600 text-white' 
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                        }`}>
                          <Database className="w-5 h-5" />
                        </div>
                        <div className="flex-1">
                          <h4 className="text-sm font-semibold text-gray-900 dark:text-white">Managed Service</h4>
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Database, cache, etc.</p>
                        </div>
                      </div>
                    </label>
                  </div>
                </div>

                {/* Basic Configuration */}
                <div className="space-y-6">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Basic Configuration</h3>
                  
                  {/* Service Name */}
                  <div className="space-y-2">
                    <label htmlFor="name" className="block text-sm font-semibold text-gray-900 dark:text-white">
                      Service Name <span className="text-red-500">*</span>
                    </label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Server className="h-5 w-5 text-gray-400" />
                      </div>
                      <input
                        type="text"
                        name="name"
                        id="name"
                        required
                        disabled={isSubmitting}
                        className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                        placeholder="Enter service name"
                      />
                    </div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Choose a descriptive name for your service
                    </p>
                  </div>

                  {/* Environment */}
                  <div className="space-y-2">
                    <label htmlFor="environmentId" className="block text-sm font-semibold text-gray-900 dark:text-white">
                      Environment <span className="text-red-500">*</span>
                    </label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Settings className="h-5 w-5 text-gray-400" />
                      </div>
                      <select
                        name="environmentId"
                        id="environmentId"
                        required
                        disabled={isSubmitting}
                        className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                      >
                        <option value="">Select environment</option>
                        {environments && environments.map((env) => (
                          <option key={env.id} value={env.id}>
                            {env.name}
                          </option>
                        ))}
                      </select>
                    </div>
                    {environments && environments.length === 0 && (
                      <p className="text-xs text-red-500">
                        No environments found. Create an environment first.
                      </p>
                    )}
                  </div>
                </div>

                {/* Git-specific Fields */}
                {serviceType === 'git' && (
                  <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Git Configuration</h3>
                    
                    {/* Repository URL */}
                    <div className="space-y-2">
                      <label htmlFor="repoUrl" className="block text-sm font-semibold text-gray-900 dark:text-white">
                        Repository URL <span className="text-red-500">*</span>
                      </label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <GitBranch className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="url"
                          name="repoUrl"
                          id="repoUrl"
                          required={serviceType === 'git'}
                          disabled={isSubmitting}
                          className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                          placeholder="https://github.com/username/repository.git"
                        />
                      </div>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        HTTPS URL of your Git repository
                      </p>
                    </div>

                    {/* Branch & Port */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-2">
                        <label htmlFor="branch" className="block text-sm font-semibold text-gray-900 dark:text-white">
                          Branch
                        </label>
                        <input
                          type="text"
                          name="branch"
                          id="branch"
                          defaultValue="main"
                          disabled={isSubmitting}
                          className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                          placeholder="main"
                        />
                      </div>

                      <div className="space-y-2">
                        <label htmlFor="port" className="block text-sm font-semibold text-gray-900 dark:text-white">
                          Port
                        </label>
                        <input
                          type="number"
                          name="port"
                          id="port"
                          defaultValue="3000"
                          disabled={isSubmitting}
                          className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                          placeholder="3000"
                        />
                      </div>
                    </div>

                    {/* Repository Visibility */}
                    <div className="space-y-3">
                      <label className="flex items-center gap-3 cursor-pointer">
                        <input
                          type="checkbox"
                          name="isPublic"
                          checked={isPublic}
                          onChange={(e) => setIsPublic(e.target.checked)}
                          disabled={isSubmitting}
                          className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                        />
                        <span className="text-sm font-semibold text-gray-900 dark:text-white flex items-center gap-1.5">
                          {isPublic ? <Globe className="h-4 w-4" /> : <Lock className="h-4 w-4" />}
                          Public repository
                        </span>
                      </label>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        Uncheck for private repositories — you'll need a personal access token (PAT).
                      </p>
                    </div>

                    {/* Private repo credentials */}
                    {!isPublic && (
                      <div className="space-y-6 rounded-xl border border-gray-200 dark:border-gray-700 p-4 bg-gray-50/30 dark:bg-gray-900/30">
                        <div className="space-y-2">
                          <label htmlFor="gitToken" className="block text-sm font-semibold text-gray-900 dark:text-white">
                            Access Token (PAT) <span className="text-red-500">*</span>
                          </label>
                          <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                              <KeyRound className="h-5 w-5 text-gray-400" />
                            </div>
                            <input
                              type="password"
                              name="gitToken"
                              id="gitToken"
                              required={!isPublic}
                              disabled={isSubmitting}
                              autoComplete="off"
                              className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                              placeholder="ghp_xxxxxxxxxxxx"
                            />
                          </div>
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            Used only to clone your repository. Stored server-side and never shown again.
                          </p>
                        </div>

                        <div className="space-y-2">
                          <label htmlFor="gitUsername" className="block text-sm font-semibold text-gray-900 dark:text-white">
                            Username <span className="text-gray-400 font-normal">(optional)</span>
                          </label>
                          <input
                            type="text"
                            name="gitUsername"
                            id="gitUsername"
                            disabled={isSubmitting}
                            autoComplete="off"
                            className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                            placeholder="Leave blank for GitHub; use 'oauth2' for GitLab"
                          />
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            Most providers (e.g. GitHub) don't need this. Required only for some like Bitbucket.
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {/* Managed Service Fields */}
                {serviceType === 'managed' && (
                  <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Service Configuration</h3>
                    
                    {/* Service Type */}
                    <div className="space-y-2">
                      <label htmlFor="managedType" className="block text-sm font-semibold text-gray-900 dark:text-white">
                        Service Type <span className="text-red-500">*</span>
                      </label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <Package className="h-5 w-5 text-gray-400" />
                        </div>
                        <select
                          name="managedType"
                          id="managedType"
                          value={managedType}
                          onChange={(e) => {
                            const selectedType = e.target.value as ManagedServiceType;
                            setManagedType(selectedType);
                            const config = selectedType ? getManagedServiceConfig(selectedType) : null;
                            if (config) {
                              setDefaultVersion(config.defaultVersion);
                            }
                          }}
                          required={serviceType === 'managed'}
                          disabled={isSubmitting}
                          className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                        >
                          <option value="">Select service type</option>
                          {Object.entries(MANAGED_SERVICE_CONFIGS).map(([key, value]) => (
                            <option key={key} value={key}>
                              {value.name} - {value.description}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>

                    {/* Version & Storage */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-2">
                        <label htmlFor="version" className="block text-sm font-semibold text-gray-900 dark:text-white">
                          Version
                        </label>
                        <input
                          type="text"
                          name="version"
                          id="version"
                          value={defaultVersion}
                          onChange={(e) => setDefaultVersion(e.target.value)}
                          disabled={isSubmitting}
                          className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                          placeholder="latest"
                        />
                        {managedType && (
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            Recommended: {getManagedServiceConfig(managedType as ManagedServiceType)?.defaultVersion}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <label htmlFor="storageSize" className="block text-sm font-semibold text-gray-900 dark:text-white">
                          Storage Size
                        </label>
                        <input
                          type="text"
                          name="storageSize"
                          id="storageSize"
                          defaultValue="1Gi"
                          ref={storageSizeRef}
                          disabled={isSubmitting}
                          onChange={(e) => {
                            const value = e.target.value;
                            // Validate format (numbers followed by Gi or Mi)
                            const sizeMatch = value.match(/^([0-9]+)([A-Za-z]+)$/); 
                            if (!sizeMatch) {
                              setStorageSizeError('Invalid format. Use 1Gi, 500Mi, etc.');
                              return;
                            }
                            
                            const numValue = parseInt(sizeMatch[1], 10);
                            const unit = sizeMatch[2].toLowerCase();
                            
                            // Check units
                            if (unit !== 'gi' && unit !== 'mi') {
                              setStorageSizeError('Only Gi or Mi units are allowed');
                              return;
                            }
                            
                            // Calculate size in Gi
                            let sizeInGi = numValue;
                            if (unit === 'mi') {
                              sizeInGi = numValue / 1024;
                            }
                            
                            // Check maximum
                            if (sizeInGi > 8) {
                              setStorageSizeError('Storage size cannot exceed 8Gi');
                            } else {
                              setStorageSizeError('');
                            }
                          }}
                          className={`w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border ${storageSizeError ? 'border-red-500' : 'border-gray-200 dark:border-gray-700'} rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50`}
                          placeholder="1Gi"
                        />
                        {storageSizeError ? (
                          <p className="text-xs text-red-500 mt-1">{storageSizeError}</p>
                        ) : (
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            Format: 1Gi, 500Mi, etc. Maximum: 8Gi
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {/* Action Buttons */}
                <div className="flex flex-col-reverse sm:flex-row gap-4 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
                  <Link
                    to={`/projects/${projectId}`}
                    className="flex-1 sm:flex-none px-6 py-3 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200 flex items-center justify-center"
                  >
                    Cancel
                  </Link>
                  
                  <button
                    type="submit"
                    disabled={isSubmitting || environments?.length === 0 || (serviceType === 'managed' && !!storageSizeError)}
                    className="flex-1 sm:flex-none px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center transform hover:scale-105 disabled:hover:scale-100"
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 size={16} className="mr-2 animate-spin" />
                        Creating Service...
                      </>
                    ) : (
                      <>
                        <Plus size={16} className="mr-2" />
                        Create Service
                      </>
                    )}
                  </button>
                </div>
              </Form>
            </div>
          </div>

          {/* Help Section */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-indigo-50/60 dark:bg-indigo-900/20 backdrop-blur-sm rounded-xl border border-indigo-200/50 dark:border-indigo-700/50 p-6">
              <h3 className="text-sm font-semibold text-indigo-800 dark:text-indigo-200 mb-3">💡 Service Types</h3>
              <div className="space-y-2 text-sm text-indigo-700 dark:text-indigo-300">
                <p><strong>Git Repository:</strong> Deploy applications from your Git repository with automatic builds</p>
                <p><strong>Managed Service:</strong> Pre-configured services like databases, caches, and message queues</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}