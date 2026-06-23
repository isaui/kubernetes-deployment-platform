import { useState } from 'react';
import { ClipboardCopy, Check, Shield, Info, GitBranch, Server, Database, Calendar, Globe, Cpu, HardDrive, Layers } from 'lucide-react';
import type { Service } from '~/types/service';

interface InformationTabProps {
  service: Service;
}

export default function InformationTab({ service }: InformationTabProps) {
  const [copiedField, setCopiedField] = useState<string | null>(null);
  
  const handleCopy = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };
  
  // Format the date
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="p-8 space-y-8">
      {/* Basic Information */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              <Info className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              <h4 className="font-semibold text-gray-900 dark:text-white">Basic Information</h4>
            </div>
          </div>
          <div className="p-6 space-y-6">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Service ID</p>
              <div className="flex items-center space-x-2">
                <p className="font-mono text-sm bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 flex-1">{service.id}</p>
                <button
                  onClick={() => handleCopy(service.id, 'id')}
                  className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                >
                  {copiedField === 'id' ? <Check className="w-4 h-4 text-green-500" /> : <ClipboardCopy className="w-4 h-4" />}
                </button>
              </div>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Name</p>
              <p className="text-gray-900 dark:text-white font-medium">{service.name}</p>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Type</p>
              <span className={`inline-flex items-center px-3 py-1 rounded-lg text-sm font-medium ${
                service.type === 'git' 
                  ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400' 
                  : 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400'
              }`}>
                {service.type === 'git' ? 'Git Repository' : 'Managed Service'}
              </span>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center">
                  <Calendar className="w-4 h-4 mr-1" />
                  Created
                </p>
                <p className="text-sm text-gray-900 dark:text-white">{formatDate(service.createdAt)}</p>
              </div>
              
              <div>
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center">
                  <Calendar className="w-4 h-4 mr-1" />
                  Updated
                </p>
                <p className="text-sm text-gray-900 dark:text-white">{formatDate(service.updatedAt)}</p>
              </div>
            </div>
          </div>
        </div>
        
        {/* Type-specific Information */}
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              {service.type === 'git' ? (
                <GitBranch className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              ) : (
                <Database className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              )}
              <h4 className="font-semibold text-gray-900 dark:text-white">
                {service.type === 'git' ? 'Repository Details' : 'Managed Service Details'}
              </h4>
            </div>
          </div>
          <div className="p-6 space-y-6">
            {service.type === 'git' ? (
              <>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Repository URL</p>
                  <div className="flex items-center space-x-2">
                    <p className="font-mono text-sm bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 flex-1 break-all">
                      {service.repoUrl}
                    </p>
                    <button
                      onClick={() => service.repoUrl && handleCopy(service.repoUrl, 'repoUrl')}
                      className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors flex-shrink-0"
                    >
                      {copiedField === 'repoUrl' ? <Check className="w-4 h-4 text-green-500" /> : <ClipboardCopy className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Branch</p>
                    <span className="inline-flex items-center px-3 py-1 rounded-lg bg-gray-100 dark:bg-gray-900/50 text-sm font-mono text-gray-900 dark:text-white">
                      {service.branch || 'main'}
                    </span>
                  </div>
                  
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Port</p>
                    <span className="inline-flex items-center px-3 py-1 rounded-lg bg-gray-100 dark:bg-gray-900/50 text-sm font-mono text-gray-900 dark:text-white">
                      {service.port || 3000}
                    </span>
                  </div>
                </div>
                
                {(service.buildCommand || service.startCommand) && (
                  <div className="space-y-4">
                    {service.buildCommand && (
                      <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Build Command</p>
                        <p className="font-mono text-sm bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white">
                          {service.buildCommand}
                        </p>
                      </div>
                    )}
                    
                    {service.startCommand && (
                      <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Start Command</p>
                        <p className="font-mono text-sm bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white">
                          {service.startCommand}
                        </p>
                      </div>
                    )}
                  </div>
                )}
              </>
            ) : (
              <>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Managed Type</p>
                  <span className="inline-flex items-center px-3 py-1 rounded-lg bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400 text-sm font-medium capitalize">
                    {service.managedType}
                  </span>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Version</p>
                    <span className="inline-flex items-center px-3 py-1 rounded-lg bg-gray-100 dark:bg-gray-900/50 text-sm font-mono text-gray-900 dark:text-white">
                      {service.version || 'latest'}
                    </span>
                  </div>
                  
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Storage Size</p>
                    <span className="inline-flex items-center px-3 py-1 rounded-lg bg-gray-100 dark:bg-gray-900/50 text-sm font-mono text-gray-900 dark:text-white">
                      {service.storageSize || '1Gi'}
                    </span>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      {/* API Key Section - Only shown for git services */}
      {service.type === 'git' && (
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-indigo-50/50 to-indigo-100/50 dark:from-indigo-900/20 dark:to-indigo-800/20">
            <div className="flex items-center space-x-2">
              <Shield className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              <h4 className="font-semibold text-gray-900 dark:text-white">API Key</h4>
            </div>
          </div>
          
          <div className="p-6">
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              Use this API key to authenticate webhook requests to this service
            </p>
            <div className="flex items-center space-x-2">
              <span className="font-mono text-sm bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 flex-1 break-all">
                {service.apiKey}
              </span>
              <button 
                onClick={() => handleCopy(service.apiKey || '', 'apiKey')}
                className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
              >
                {copiedField === 'apiKey' ? <Check className="w-4 h-4 text-green-500" /> : <ClipboardCopy className="w-4 h-4" />}
              </button>
            </div>
            
            <div className="mt-4 p-4 bg-yellow-50/80 dark:bg-yellow-900/20 border border-yellow-200/50 dark:border-yellow-800/50 rounded-xl">
              <p className="text-sm text-yellow-800 dark:text-yellow-200 font-medium mb-1">Security Notice</p>
              <p className="text-sm text-yellow-700 dark:text-yellow-300">
                This API key grants access to trigger deployments and check deployment status for this service.
                Treat it like a password and do not share it publicly.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Resources Section */}
      <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
          <div className="flex items-center space-x-2">
            <Server className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <h4 className="font-semibold text-gray-900 dark:text-white">Resources & Scaling</h4>
          </div>
        </div>
        
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="text-center">
              <div className="mx-auto w-12 h-12 bg-gradient-to-br from-blue-100 to-blue-200 dark:from-blue-900/30 dark:to-blue-800/30 rounded-xl flex items-center justify-center mb-3">
                <Cpu className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">CPU Limit</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{service.cpuLimit || 'Not set'}</p>
            </div>
            
            <div className="text-center">
              <div className="mx-auto w-12 h-12 bg-gradient-to-br from-green-100 to-green-200 dark:from-green-900/30 dark:to-green-800/30 rounded-xl flex items-center justify-center mb-3">
                <HardDrive className="w-6 h-6 text-green-600 dark:text-green-400" />
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Memory Limit</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{service.memoryLimit || 'Not set'}</p>
            </div>
            
            <div className="text-center">
              <div className="mx-auto w-12 h-12 bg-gradient-to-br from-purple-100 to-purple-200 dark:from-purple-900/30 dark:to-purple-800/30 rounded-xl flex items-center justify-center mb-3">
                <Layers className="w-6 h-6 text-purple-600 dark:text-purple-400" />
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Replicas</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">
                {service.isStaticReplica ? `${service.replicas}` : `auto`}
              </p>
            </div>
            
            {!service.isStaticReplica && (
              <div className="text-center">
                <div className="mx-auto w-12 h-12 bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/30 dark:to-orange-800/30 rounded-xl flex items-center justify-center mb-3">
                  <Server className="w-6 h-6 text-orange-600 dark:text-orange-400" />
                </div>
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Scaling Range</p>
                <p className="text-lg font-semibold text-gray-900 dark:text-white">{service.minReplicas} - {service.maxReplicas}</p>
              </div>
            )}
          </div>
        </div>
      </div>
      
      {/* Domain Information */}
      {(service.domain || service.customDomain) && (
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              <Globe className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              <h4 className="font-semibold text-gray-900 dark:text-white">Domain Information</h4>
            </div>
          </div>
          
          <div className="p-6 space-y-6">
            {service.domain && (
              <div>
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Default Domain</p>
                <div className="flex items-center space-x-2">
                  <a 
                    href={`https://${service.domain}`} 
                    target="_blank" 
                    rel="noopener noreferrer" 
                    className="font-mono text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 flex-1 break-all transition-colors"
                  >
                    {service.domain}
                  </a>
                  <button
                    onClick={() => handleCopy(service.domain || '', 'domain')}
                    className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                  >
                    {copiedField === 'domain' ? <Check className="w-4 h-4 text-green-500" /> : <ClipboardCopy className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            )}
            
            {service.customDomain && (
              <div>
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Custom Domain</p>
                <div className="flex items-center space-x-2">
                  <a 
                    href={`https://${service.customDomain}`} 
                    target="_blank" 
                    rel="noopener noreferrer" 
                    className="font-mono text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 bg-gray-100 dark:bg-gray-900/50 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 flex-1 break-all transition-colors"
                  >
                    {service.customDomain}
                  </a>
                  <button
                    onClick={() => handleCopy(service.customDomain || '', 'customDomain')}
                    className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                  >
                    {copiedField === 'customDomain' ? <Check className="w-4 h-4 text-green-500" /> : <ClipboardCopy className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}