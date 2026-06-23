import { useState, useEffect } from 'react';
import { Clock, GitCommit, CheckCircle, XCircle, AlertCircle, RefreshCw, Rocket, Activity, Loader2 } from 'lucide-react';
import type { Service } from '~/types/service';
import type { Deployment } from '~/types/deployment';
import { createDeployment, getServiceDeployments } from '~/actions/deployment';

interface DeploymentTabProps {
  service: Service;
}

export default function DeploymentTab({ service }: DeploymentTabProps) {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isDeploying, setIsDeploying] = useState(false);
  const [isDeployingLatest, setIsDeployingLatest] = useState(false);
  
  const fetchDeployments = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await getServiceDeployments(service.id);
      setDeployments(response || []);
    } catch (err) {
      setError((err as Error).message || 'Failed to load deployments');
    } finally {
      setIsLoading(false);
    }
  };
  
  // Fetch deployments
  useEffect(() => {
    if (!service.id) return;
    
    fetchDeployments();
  }, [service.id]);
  
  const redeploy = async (deployment: Deployment) => {
    if (isDeploying || deployment.status === 'building') return;
    
    setIsDeploying(true);
    try {
      await createDeployment(service.id, {
        commitId: deployment.commitSha,
        commitMessage: deployment.commitMessage,
        apiKey: service.apiKey,
      });
      await new Promise(resolve => setTimeout(resolve, 1000));
      fetchDeployments();
    } catch (err) {
      setError((err as Error).message || 'Failed to trigger deployment');
    } finally {
      setIsDeploying(false);
    }
  };

  // Deploy latest version (without specific commit)
  const deployLatest = async () => {
    if (isDeploying || isDeployingLatest) return;
    
    setIsDeployingLatest(true);
    try {
      // Deploy latest version with API key but without commitId
      await createDeployment(service.id, {
        apiKey: service.apiKey,
      });
      await new Promise(resolve => setTimeout(resolve, 1000));
      fetchDeployments();
    } catch (err) {
      setError((err as Error).message || 'Failed to deploy latest version');
    } finally {
      setIsDeployingLatest(false);
    }
  };
  
  // Function to format date for better readability
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };
  
  // Function to get status icon
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'failed':
        return <XCircle className="w-5 h-5 text-red-500" />;
      case 'building':
        return <RefreshCw className="w-5 h-5 text-yellow-500 animate-spin" />;
      default:
        return <AlertCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  // Function to get status badge styling
  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-100/80 text-green-800 dark:bg-green-900/30 dark:text-green-400 border border-green-200/50 dark:border-green-700/50';
      case 'failed':
        return 'bg-red-100/80 text-red-800 dark:bg-red-900/30 dark:text-red-400 border border-red-200/50 dark:border-red-700/50';
      case 'building':
        return 'bg-yellow-100/80 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 border border-yellow-200/50 dark:border-yellow-700/50';
      default:
        return 'bg-gray-100/80 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400 border border-gray-200/50 dark:border-gray-700/50';
    }
  };
  
  return (
    <div className="p-8 space-y-8">


      {/* Action Buttons */}
      <div className="flex justify-center">
        <div className="bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 p-4 shadow-sm">
          <div className="flex flex-wrap gap-3 justify-center">
            {service.type === 'git' && (
              <button
                onClick={deployLatest}
                disabled={isDeploying || isDeployingLatest}
                className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105 disabled:opacity-50 disabled:hover:scale-100 disabled:cursor-not-allowed"
              >
                {isDeployingLatest ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Deploying...
                  </>
                ) : (
                  <>
                    <GitCommit className="w-4 h-4 mr-2" />
                    Deploy Latest
                  </>
                )}
              </button>
            )}
            
            <button
              onClick={fetchDeployments}
              disabled={isLoading}
              className="inline-flex items-center px-6 py-3 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200 disabled:opacity-50"
            >
              <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-6 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <XCircle className="w-6 h-6 text-red-500 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="text-sm font-semibold text-red-800 dark:text-red-200 mb-1">Deployment Error</h3>
              <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
            </div>
          </div>
        </div>
      )}
      
      {/* Loading State */}
      {isLoading ? (
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-12 text-center">
          <div className="flex items-center justify-center space-x-3">
            <Loader2 className="h-8 w-8 text-indigo-600 animate-spin" />
            <span className="text-lg font-medium text-gray-900 dark:text-white">Loading deployments...</span>
          </div>
        </div>
      ) : deployments.length === 0 ? (
        /* Empty State */
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-12 text-center">
          <div className="mx-auto w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-2xl flex items-center justify-center mb-4">
            <GitCommit className="w-8 h-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">No Deployments</h3>
          <p className="text-gray-500 dark:text-gray-400 mb-6">
            {service.type === 'git' 
              ? 'Start by deploying your application.' 
              : 'This managed service has no deployment history yet.'}
          </p>
          {service.type === 'git' && (
            <button
              onClick={deployLatest}
              disabled={isDeployingLatest}
              className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105"
            >
              <GitCommit className="w-4 h-4 mr-2" />
              Deploy Now
            </button>
          )}
        </div>
      ) : (
        /* Deployments Table */
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              <Activity className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Recent Deployments</h3>
              <span className="inline-flex items-center px-2 py-1 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-xs font-medium text-indigo-700 dark:text-indigo-300">
                {deployments.length}
              </span>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead className="bg-gray-50/50 dark:bg-gray-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Commit</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Message</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Date</th>
                  <th className="px-6 py-3 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200/50 dark:divide-gray-700/50">
                {deployments.map((deployment) => (
                  <tr key={deployment.id} className="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center space-x-3">
                        {getStatusIcon(deployment.status)}
                        <span className={`inline-flex items-center px-3 py-1 rounded-lg text-xs font-medium backdrop-blur-sm ${getStatusBadgeClass(deployment.status)}`}>
                          {deployment.status.charAt(0).toUpperCase() + deployment.status.slice(1)}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {deployment.commitSha ? (
                        <div className="text-sm font-mono bg-gray-100 dark:bg-gray-900/50 px-3 py-1 rounded-lg text-gray-900 dark:text-gray-200 inline-block">
                          {deployment.commitSha.substring(0, 7)}
                        </div>
                      ) : (
                        <span className="text-sm text-gray-500 dark:text-gray-400">N/A</span>
                      )}
                    </td>
                    <td className="px-6 py-4 max-w-xs">
                      <div className="text-sm text-gray-900 dark:text-gray-200 truncate">
                        {deployment.commitMessage || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
                        <Clock className="flex-shrink-0 mr-2 h-4 w-4" />
                        {formatDate(deployment.createdAt)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      {service.type === 'git' && deployment.status !== 'building' && (
                        <button 
                          onClick={() => redeploy(deployment)}
                          disabled={isDeploying}
                          className="inline-flex items-center px-3 py-1.5 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 font-medium rounded-lg hover:bg-indigo-200 dark:hover:bg-indigo-900/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-xs"
                        >
                          {isDeploying ? 'Deploying...' : 'Redeploy'}
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}