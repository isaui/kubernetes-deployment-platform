import { ArrowUpRight, CheckCircle, Clock, Edit, Server, Trash2, XCircle, AlertTriangle, Eye, Activity, Database, TrendingUp } from 'lucide-react';
import { Link } from '@remix-run/react';
import { useState } from 'react';
import type { ProjectStatsResponse, ProjectEnvironmentItem } from '~/types/project';
import { deleteEnvironment } from '~/actions/environment';

interface ProjectStatsProps {
  stats: ProjectStatsResponse;
  onEditEnvironment?: (environment: ProjectEnvironmentItem) => void;
  onDeleteSuccess?: () => void;
}

export default function ProjectStats({ stats, onEditEnvironment, onDeleteSuccess = () => window.location.reload() }: ProjectStatsProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [environmentToDelete, setEnvironmentToDelete] = useState<ProjectEnvironmentItem | null>(null);
  
  // Handle environment deletion
  const handleDelete = async (env: ProjectEnvironmentItem) => {
    setEnvironmentToDelete(env);
  };
  
  const confirmDelete = async () => {
    if (!environmentToDelete) return;
    
    setIsDeleting(true);
    setDeleteError(null);
    
    try {
      await deleteEnvironment(environmentToDelete.id);
      setEnvironmentToDelete(null);
      onDeleteSuccess();
    } catch (error) {
      setDeleteError(error instanceof Error ? error.message : 'Failed to delete environment');
    } finally {
      setIsDeleting(false);
    }
  };
  
  const cancelDelete = () => {
    setEnvironmentToDelete(null);
    setDeleteError(null);
  };

  // Function to create a color class based on service status
  const getStatusColorClass = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'running':
      case 'active':
        return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300';
      case 'pending':
      case 'building':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300';
      case 'failed':
      case 'error':
        return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300';
    }
  };

  // Format success rate as percentage
  const formatSuccessRate = (rate: number): string => {
    return `${(rate * 100).toFixed(1)}%`;
  };

  return (
    <div className="space-y-8">
      {/* Delete Confirmation Modal */}
      {environmentToDelete && (
        <>
          <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40" />
          <div className="fixed inset-0 flex items-center justify-center z-50 p-4">
            <div className="bg-white/95 dark:bg-gray-800/95 backdrop-blur-md rounded-2xl shadow-xl border border-gray-200/50 dark:border-gray-700/50 w-full max-w-md">
              <div className="p-6">
                {/* Icon */}
                <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 dark:bg-red-900/30 mb-4">
                  <AlertTriangle className="h-8 w-8 text-red-600 dark:text-red-400" />
                </div>

                {/* Title & Message */}
                <div className="text-center space-y-3 mb-6">
                  <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                    Delete Environment
                  </h3>
                  <p className="text-gray-600 dark:text-gray-300">
                    Are you sure you want to delete <span className="font-semibold text-gray-900 dark:text-white">"{environmentToDelete.name}"</span>?
                  </p>
                  <p className="text-sm text-red-600 dark:text-red-400 font-medium">
                    This action cannot be undone.
                  </p>
                </div>

                {deleteError && (
                  <div className="mb-4 rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-3">
                    <p className="text-sm text-red-700 dark:text-red-300">{deleteError}</p>
                  </div>
                )}

                {/* Actions */}
                <div className="flex flex-col-reverse sm:flex-row gap-3">
                  <button
                    onClick={cancelDelete}
                    disabled={isDeleting}
                    className="flex-1 px-4 py-3 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors disabled:opacity-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={confirmDelete}
                    disabled={isDeleting}
                    className="flex-1 px-4 py-3 bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800 text-white font-semibold rounded-xl transition-all duration-200 disabled:opacity-50 flex items-center justify-center"
                  >
                    {isDeleting ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2"></div>
                        Deleting...
                      </>
                    ) : (
                      'Delete Environment'
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Services</p>
              <h3 className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.services.total}</h3>
            </div>
            <div className="p-3 bg-gradient-to-br from-blue-100 to-blue-200 dark:from-blue-900/30 dark:to-blue-800/30 rounded-xl">
              <Server className="h-8 w-8 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </div>

        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Git Deployments</p>
              <h3 className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.deployments.total}</h3>
            </div>
            <div className="p-3 bg-gradient-to-br from-purple-100 to-purple-200 dark:from-purple-900/30 dark:to-purple-800/30 rounded-xl">
              <ArrowUpRight className="h-8 w-8 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </div>

        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Git Success Rate</p>
              <h3 className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">
                {formatSuccessRate(stats.deployments.successRate)}
              </h3>
            </div>
            <div className="p-3 bg-gradient-to-br from-green-100 to-green-200 dark:from-green-900/30 dark:to-green-800/30 rounded-xl">
              <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </div>

        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">In Progress</p>
              <h3 className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">
                {stats.deployments.inProgress}
              </h3>
            </div>
            <div className="p-3 bg-gradient-to-br from-yellow-100 to-yellow-200 dark:from-yellow-900/30 dark:to-yellow-800/30 rounded-xl">
              <Clock className="h-8 w-8 text-yellow-600 dark:text-yellow-400" />
            </div>
          </div>
        </div>
      </div>

      {/* Environments Section */}
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
          <div className="flex items-center space-x-2">
            <Database className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Environments</h3>
            <span className="inline-flex items-center px-2 py-1 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-xs font-medium text-indigo-700 dark:text-indigo-300">
              {stats.environments.total}
            </span>
          </div>
        </div>
        
        <div className="overflow-x-auto">
          <div className="min-w-full inline-block align-middle">
            <table className="min-w-full">
              <thead className="bg-gray-50/50 dark:bg-gray-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap">Name</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap">Description</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap">Services</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap">Created</th>
                  <th className="px-6 py-3 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap w-24">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200/50 dark:divide-gray-700/50">
                {stats.environments.environments && stats.environments.environments.length > 0 ? (
                  stats.environments.environments.map((env) => (
                    <tr key={env.id} className="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm font-semibold text-gray-900 dark:text-white">{env.name}</span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-sm text-gray-600 dark:text-gray-400 max-w-xs truncate block">{env.description || '-'}</span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="inline-flex items-center px-2 py-1 rounded-lg bg-blue-100 dark:bg-blue-900/30 text-xs font-medium text-blue-700 dark:text-blue-300">
                          {env.servicesCount || 0} services
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {new Date(env.createdAt).toLocaleDateString()}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap w-24">
                        <div className="flex items-center justify-end space-x-1">
                          {onEditEnvironment && (
                            <button 
                              onClick={() => onEditEnvironment(env)}
                              className="p-2 text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                              title="Edit environment"
                            >
                              <Edit className="h-4 w-4" />
                            </button>
                          )}
                          
                          {env.servicesCount === 0 && (
                            <button 
                              onClick={() => handleDelete(env)}
                              className="p-2 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                              title="Delete environment"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center">
                      <div className="space-y-3">
                        <Database className="mx-auto h-12 w-12 text-gray-400" />
                        <p className="text-gray-500 dark:text-gray-400">No environments created yet</p>
                        <p className="text-sm text-gray-400 dark:text-gray-500">Create an environment to organize your services</p>
                      </div>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Services Section */}
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
          <div className="flex items-center space-x-2">
            <Activity className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Services</h3>
            <span className="inline-flex items-center px-2 py-1 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-xs font-medium text-indigo-700 dark:text-indigo-300">
              {stats.services.total}
            </span>
          </div>
        </div>
        
        <div className="overflow-x-auto">
          <table className="min-w-full">
            <thead className="bg-gray-50/50 dark:bg-gray-800/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Environment</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Replicas</th>
                
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200/50 dark:divide-gray-700/50">
              {stats.services.servicesList && stats.services.servicesList.length > 0 ? (
                stats.services.servicesList.map((service) => (
                  <tr key={service.id} className="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
                    <td className="px-6 py-4">
                      <Link 
                        to={`/services/${service.id}`} 
                        className="text-sm font-semibold text-gray-900 dark:text-white hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                      >
                        {service.name}
                      </Link>
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center px-2 py-1 rounded-lg bg-gray-100 dark:bg-gray-800 text-xs font-medium text-gray-800 dark:text-gray-200">
                        {service.type}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center px-2 py-1 rounded-lg bg-blue-100 dark:bg-blue-800 text-xs font-medium text-blue-800 dark:text-blue-200">
                        {service.environmentName || 'Not Set'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`inline-flex items-center px-2 py-1 rounded-lg text-xs font-medium ${getStatusColorClass(service.status)}`}>
                        {service.status}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      {service.isAutoScaling ? (
                        <div className="flex items-center space-x-2">
                          <span className="inline-flex items-center px-1.5 py-0.5 rounded bg-blue-100 dark:bg-blue-900/30 text-xs font-medium text-blue-800 dark:text-blue-300">
                            Auto
                          </span>
                        </div>
                      ) : (
                        <span className="text-sm text-gray-900 dark:text-white">{service.replicas}</span>
                      )}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center">
                    <div className="space-y-3">
                      <Server className="mx-auto h-12 w-12 text-gray-400" />
                      <p className="text-gray-500 dark:text-gray-400">No services deployed yet</p>
                      <p className="text-sm text-gray-400 dark:text-gray-500">Deploy your first service to get started</p>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Deployment Statistics */}
      {stats.deployments.total > 0 && (
        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
            <div className="flex items-center space-x-2">
              <TrendingUp className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Deployment Statistics</h3>
            </div>
          </div>
          
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex items-center space-x-4">
                <div className="p-3 bg-gradient-to-br from-green-100 to-green-200 dark:from-green-900/30 dark:to-green-800/30 rounded-xl">
                  <CheckCircle className="h-6 w-6 text-green-600 dark:text-green-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Successful</p>
                  <h4 className="text-2xl font-bold text-gray-900 dark:text-white">{stats.deployments.successful}</h4>
                </div>
              </div>
              
              <div className="flex items-center space-x-4">
                <div className="p-3 bg-gradient-to-br from-red-100 to-red-200 dark:from-red-900/30 dark:to-red-800/30 rounded-xl">
                  <XCircle className="h-6 w-6 text-red-600 dark:text-red-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Failed</p>
                  <h4 className="text-2xl font-bold text-gray-900 dark:text-white">{stats.deployments.failed}</h4>
                </div>
              </div>
              
              <div className="flex items-center space-x-4">
                <div className="p-3 bg-gradient-to-br from-yellow-100 to-yellow-200 dark:from-yellow-900/30 dark:to-yellow-800/30 rounded-xl">
                  <Clock className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">In Progress</p>
                  <h4 className="text-2xl font-bold text-gray-900 dark:text-white">{stats.deployments.inProgress}</h4>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}