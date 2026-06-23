import  { useState } from 'react';
import type { 
  ClusterInfo, 
  NodeStats, 
  PodStats, 
  DeploymentStats,
  ServiceStats,
  IngressStats,
  CertificateStats,
  PVCStats
} from '~/types/stats';
import type { Registry } from '~/types/registry';

// Import tab components
import ClusterOverview from './ClusterOverview';
import NodesOverview from './NodesOverview';
import WorkloadsOverview from './WorkloadsOverview';
import NetworkingOverview from './NetworkingOverview';
import SecurityOverview from './SecurityOverview';
import RegistriesOverview from './RegistriesOverview';
import PVCOverview from './PVCOverview';

interface AdminDashboardProps {
  clusterInfo: ClusterInfo;
  nodes: NodeStats[];
  pods: PodStats[];
  deployments: DeploymentStats[];
  services: ServiceStats[];
  ingresses: IngressStats[];
  certificates: CertificateStats[];
  registries: Registry[];
  pvcs: PVCStats[];
}

export default function AdminDashboard({ 
  clusterInfo,
  nodes,
  pods,
  deployments,
  services,
  ingresses,
  certificates,
  registries,
  pvcs
}: AdminDashboardProps) {
  // Define available tabs
  const tabs = [
    { id: 'overview', name: 'Overview' },
    { id: 'nodes', name: 'Nodes' },
    { id: 'workloads', name: 'Workloads' },
    { id: 'networking', name: 'Networking' },
    { id: 'storage', name: 'Storage' },
    { id: 'security', name: 'Security' },
    { id: 'registries', name: 'Registries' }
  ];
  
  // State for current active tab
  const [currentTab, setCurrentTab] = useState<string>('overview');

  return (
    <div className="bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        <div className='flex items-center justify-between gap-x-2 mb-6'>
            <h1 className="text-2xl font-bold text-gray-900">Kubernetes Admin Dashboard</h1>  
            <a
                href="/projects"
                className={` flex-shrink-0inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md 
                  bg-indigo-600 text-white hover:bg-indigo-700`}
              >
                Project Deployments
              </a>
        </div>
        
        
        {/* Tab navigation */}
        <div className="border-b border-gray-200 mb-6">
          <div className="-mb-px flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setCurrentTab(tab.id)}
                className={`
                  whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm
                  ${
                    currentTab === tab.id
                      ? 'border-indigo-500 text-indigo-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }
                `}
              >
                {tab.name}
              </button>
            ))}
          </div>
        </div>
        
        {/* Tab content */}
        <div>
          {currentTab === 'overview' && (
            <ClusterOverview clusterInfo={clusterInfo} />
          )}
          
          {currentTab === 'nodes' && (
            <NodesOverview nodes={nodes} />
          )}
          
          {currentTab === 'workloads' && (
            <WorkloadsOverview pods={pods} deployments={deployments} />
          )}
          
          {currentTab === 'networking' && (
            <NetworkingOverview services={services} ingresses={ingresses} />
          )}
          
          {currentTab === 'storage' && (
            <PVCOverview pvcs={pvcs} />
          )}
          
          {currentTab === 'security' && <SecurityOverview certificates={certificates} />}
          {currentTab === 'registries' && <RegistriesOverview initialRegistries={registries} />}
        </div>
      </div>
    </div>
  );
}
