import { useState } from 'react';
import type { DeploymentStats, PodStats } from '~/types/stats';
import { DeploymentsList } from './workloads/DeploymentsList';
import { PodsList } from './workloads/PodsList';

interface WorkloadsOverviewProps {
  deployments: DeploymentStats[];
  pods: PodStats[];
}

export default function WorkloadsOverview({ deployments, pods }: WorkloadsOverviewProps) {
  const [activeWorkloadView, setActiveWorkloadView] = useState<'deployments' | 'pods'>('deployments');

  return (
    <div className="space-y-6">
      {/* Workload type selector */}
      <div className="bg-white shadow rounded-lg p-4">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-xl font-semibold text-gray-900">Workloads</h1>
            <p className="mt-2 text-sm text-gray-700">
              View and manage deployments and pods across your cluster
            </p>
          </div>
          <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
            <div className="flex space-x-3">
              <button
                type="button"
                onClick={() => setActiveWorkloadView('deployments')}
                className={`inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md 
                  ${activeWorkloadView === 'deployments' 
                    ? 'bg-indigo-600 text-white hover:bg-indigo-700' 
                    : 'bg-white text-gray-700 hover:bg-gray-50 border-gray-300'
                  }`}
              >
                Deployments
              </button>
              <button
                type="button"
                onClick={() => setActiveWorkloadView('pods')}
                className={`inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md 
                  ${activeWorkloadView === 'pods' 
                    ? 'bg-indigo-600 text-white hover:bg-indigo-700' 
                    : 'bg-white text-gray-700 hover:bg-gray-50 border-gray-300'
                  }`}
              >
                Pods
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Workload summary */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-medium mb-4">Workloads Summary</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Total Deployments</h3>
            <p className="mt-1 text-lg font-semibold">{deployments.length}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Total Pods</h3>
            <p className="mt-1 text-lg font-semibold">{pods.length}</p>
          </div>
        </div>
      </div>

      {/* Display appropriate component based on active view */}
      {activeWorkloadView === 'deployments' ? (
        <DeploymentsList deployments={deployments} />
      ) : (
        <PodsList pods={pods} />
      )}
    </div>
  );
}
