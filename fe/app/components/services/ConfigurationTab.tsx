import { useState } from 'react';
import { RefreshCw, Save, Settings, Server, GitBranch, Database, AlertCircle, Loader2, Globe, Cpu, HardDrive, Layers } from 'lucide-react';
import type { Service } from '~/types/service';
import { createServiceUpdate } from '~/types/service-update';
import { updateService } from '~/actions/service-update';

interface ConfigurationTabProps {
  service: Service;
  onServiceUpdated?: () => void;
}

export default function ConfigurationTab({ service, onServiceUpdated }: ConfigurationTabProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  
  const [formValues, setFormValues] = useState({
    name: service.name || '',
    cpuLimit: service.cpuLimit || '',
    memoryLimit: service.memoryLimit || '',
    // FIX: Handle boolean properly - use ?? instead of ||
    isStaticReplica: service.isStaticReplica ?? false,
    replicas: service.replicas,
    minReplicas: service.minReplicas,
    maxReplicas: service.maxReplicas,
    customDomain: service.customDomain || '',
    // Git service fields (removed buildCommand and startCommand)
    branch: service.branch || '',
    port: service.port ? String(service.port) : '',
    // Domain fields
    domain: service.domain || '',
    // Managed service fields
    version: service.version || '',
    storageSize: service.storageSize || ''
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    let processedValue: any = value;
    
    // Handle checkbox for isStaticReplica
    if (type === 'checkbox') {
      processedValue = (e.target as HTMLInputElement).checked;
    }
    // Handle numeric inputs
    else if (name === 'port' || name === 'replicas' || name === 'minReplicas' || name === 'maxReplicas') {
      processedValue = value === '' ? '' : Number(value);
    }
    
    setFormValues(prev => ({ ...prev, [name]: processedValue }));
  };

  // FIX: Custom comparison function for detecting changes
  const hasFieldChanged = (fieldName: string): boolean => {
    const formValue = formValues[fieldName as keyof typeof formValues];
    const serviceValue = service[fieldName as keyof Service];
    
    // Special handling for boolean fields
    if (fieldName === 'isStaticReplica') {
      // Normalize both values to handle undefined/null cases
      const normalizedFormValue = Boolean(formValue);
      const normalizedServiceValue = Boolean(serviceValue);
      return normalizedFormValue !== normalizedServiceValue;
    }
    
    // Special handling for port (number vs string)
    if (fieldName === 'port') {
      const formPort = formValue === '' ? undefined : Number(formValue);
      return formPort !== serviceValue;
    }
    
    // For empty strings vs undefined/null, treat as same
    if (formValue === '' && (serviceValue === undefined || serviceValue === null)) {
      return false;
    }
    if ((formValue === undefined || formValue === null) && serviceValue === '') {
      return false;
    }
    
    return formValue !== serviceValue;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    setSuccess(false);

    try {
      // FIX: Use custom comparison function
      const changedFields = Object.keys(formValues).filter(hasFieldChanged);
      
      if (changedFields.length === 0) {
        setError('No changes to save');
        setIsLoading(false);
        return;
      }

      console.log('Changed fields:', changedFields); // Debug log

      // Create update DTO with changed fields only
      // Convert form values to appropriate types for Service object
      const serviceData = {
        ...service,
        ...formValues,
        // Ensure port is a number when passed to createServiceUpdate
        port: formValues.port === '' ? undefined : Number(formValues.port),
      };
      
      const serviceUpdate = createServiceUpdate(
        serviceData,
        changedFields
      );
      
      console.log('Service update payload:', serviceUpdate); // Debug log
      
      await updateService(service.id, serviceUpdate);
      setSuccess(true);
      
      if (onServiceUpdated) {
        onServiceUpdated();
      }

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(false), 3000);
    } catch (error) {
      console.error('Failed to update service', error);
      setError(error instanceof Error ? error.message : 'Failed to update service');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="p-8 space-y-8">
      

      {/* Status Messages */}
      {error && (
        <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-4 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <AlertCircle className="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Configuration Error</h3>
              <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
            </div>
          </div>
        </div>
      )}

      {success && (
        <div className="bg-green-50/80 dark:bg-green-900/20 border border-green-200/60 dark:border-green-800/30 rounded-xl p-4 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <div className="w-5 h-5 rounded-full bg-green-500 flex items-center justify-center mt-0.5 flex-shrink-0">
              <span className="text-white text-xs">✓</span>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-green-800 dark:text-green-200">Configuration Updated</h3>
              <p className="text-sm text-green-700 dark:text-green-300 mt-1">Your service configuration has been saved successfully.</p>
            </div>
          </div>
        </div>
      )}

      {/* Main Configuration Form */}
      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Common Settings */}
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              <Server className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">General Settings</h3>
            </div>
          </div>
          
          <div className="p-6 space-y-6">
            {/* Service Name */}
            <div>
              <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="name">
                Service Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="name"
                name="name"
                value={formValues.name}
                onChange={handleChange}
                className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                placeholder="Enter service name"
                required
              />
            </div>

            {/* Resource Limits */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="cpuLimit">
                  <div className="flex items-center space-x-2">
                    <Cpu className="w-4 h-4" />
                    <span>CPU Limit</span>
                  </div>
                </label>
                <input
                  type="text"
                  id="cpuLimit"
                  name="cpuLimit"
                  value={formValues.cpuLimit}
                  onChange={handleChange}
                  placeholder="e.g., 100m, 0.5"
                  className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">CPU cores allocated to your service</p>
              </div>
              
              <div>
                <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="memoryLimit">
                  <div className="flex items-center space-x-2">
                    <HardDrive className="w-4 h-4" />
                    <span>Memory Limit</span>
                  </div>
                </label>
                <input
                  type="text"
                  id="memoryLimit"
                  name="memoryLimit"
                  value={formValues.memoryLimit}
                  onChange={handleChange}
                  placeholder="e.g., 128Mi, 1Gi"
                  className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Memory allocated to your service</p>
              </div>
            </div>

            {/* Domains */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="domain">
                  <div className="flex items-center space-x-2">
                    <Globe className="w-4 h-4" />
                    <span>Auto-generated Domain</span>
                  </div>
                </label>
                <input
                  type="text"
                  id="domain"
                  name="domain"
                  value={formValues.domain}
                  disabled={true}
                  readOnly={true}
                  className="w-full px-4 py-3 bg-gray-100/50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-500 dark:text-gray-400 backdrop-blur-sm cursor-not-allowed"
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Auto-generated domain for your service</p>
              </div>
              
              <div>
                <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="customDomain">
                  <div className="flex items-center space-x-2">
                    <Globe className="w-4 h-4" />
                    <span>Custom Domain</span>
                  </div>
                </label>
                <input
                  type="text"
                  id="customDomain"
                  name="customDomain"
                  value={formValues.customDomain}
                  onChange={handleChange}
                  placeholder="app.example.com"
                  className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                />
                <div className="mt-1 space-y-1">
                  <p className="text-xs text-gray-500 dark:text-gray-400">Optional custom domain for your service</p>
                  {formValues.customDomain && (
                    <div className="flex items-center space-x-2 bg-blue-50 dark:bg-blue-900/30 border border-blue-100 dark:border-blue-800/50 rounded-lg p-2">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4 text-blue-500 flex-shrink-0">
                        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clipRule="evenodd" />
                      </svg>
                      <div className="text-xs text-blue-700 dark:text-blue-300">
                        <p><strong>DNS Setup Required:</strong> Create an A record for <span className="font-mono bg-blue-100 dark:bg-blue-800/50 px-1 rounded">{formValues.customDomain}</span> that points to <span className="font-mono bg-blue-100 dark:bg-blue-800/50 px-1 rounded">{window.ENV?.LOAD_BALANCER_IP || "your Load Balancer IP"}</span></p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Scaling Configuration - Only shown for git services */}
        {service.type === 'git' && (
          <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
              <div className="flex items-center space-x-2">
                <Layers className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Scaling Configuration</h3>
              </div>
            </div>
            
            <div className="p-6 space-y-6">
              {/* Static Replicas Toggle */}
              <div className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  id="isStaticReplica"
                  name="isStaticReplica"
                  checked={formValues.isStaticReplica}
                  onChange={handleChange}
                  className="w-5 h-5 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500 focus:ring-2"
                />
                <label className="text-sm font-semibold text-gray-900 dark:text-white" htmlFor="isStaticReplica">
                  Static Replicas (Disable Auto-scaling)
                </label>
              </div>
              
              {formValues.isStaticReplica ? (
                <div>
                  <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="replicas">
                    Replica Count
                  </label>
                  <input
                    type="number"
                    id="replicas"
                    name="replicas"
                    value={formValues.replicas}
                    onChange={handleChange}
                    min="1"
                    className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Fixed number of replicas to run</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="minReplicas">
                      Min Replicas
                    </label>
                    <input
                      type="number"
                      id="minReplicas"
                      name="minReplicas"
                      value={formValues.minReplicas}
                      onChange={handleChange}
                      min="1"
                      className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                    />
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Minimum number of replicas</p>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="maxReplicas">
                      Max Replicas
                    </label>
                    <input
                      type="number"
                      id="maxReplicas"
                      name="maxReplicas"
                      value={formValues.maxReplicas}
                      onChange={handleChange}
                      min="1"
                      className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                    />
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Maximum number of replicas</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
        
        {/* Service-specific Settings */}
        {service.type === 'git' && (
          <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
              <div className="flex items-center space-x-2">
                <GitBranch className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Git Repository Settings</h3>
              </div>
            </div>
            
            <div className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="branch">
                    Branch
                  </label>
                  <input
                    type="text"
                    id="branch"
                    name="branch"
                    value={formValues.branch}
                    onChange={handleChange}
                    placeholder="main"
                    className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Git branch to deploy from</p>
                </div>
                
                <div>
                  <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="port">
                    Port
                  </label>
                  <input
                    type="number"
                    id="port"
                    name="port"
                    value={formValues.port}
                    onChange={handleChange}
                    placeholder="3000"
                    min="1"
                    max="65535"
                    className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Port your application runs on</p>
                </div>
              </div>
            </div>
          </div>
        )}
        
        {service.type === 'managed' && (
          <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
              <div className="flex items-center space-x-2">
                <Database className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Managed Service Settings</h3>
              </div>
            </div>
            
            <div className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="version">
                    Version
                  </label>
                  <input
                    type="text"
                    id="version"
                    name="version"
                    value={formValues.version}
                    onChange={handleChange}
                    placeholder="latest"
                    className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Service version to deploy</p>
                </div>
                
                <div>
                  <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2" htmlFor="storageSize">
                    Storage Size
                  </label>
                  <input
                    type="text"
                    id="storageSize"
                    name="storageSize"
                    value={formValues.storageSize}
                    onChange={handleChange}
                    placeholder="1Gi"
                    className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Persistent storage allocation</p>
                </div>
              </div>
            </div>
          </div>
        )}
        
        {/* Submit Button */}
        <div className="flex justify-center pt-6">
          <button
            type="submit"
            disabled={isLoading}
            className="px-8 py-4 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center transform hover:scale-105 disabled:hover:scale-100"
          >
            {isLoading ? (
              <>
                <Loader2 size={18} className="mr-3 animate-spin" />
                Saving Configuration...
              </>
            ) : (
              <>
                <Save size={18} className="mr-3" />
                Save Configuration
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}