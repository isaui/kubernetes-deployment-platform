import { useState, useEffect } from 'react';
import { json, redirect, type LoaderFunctionArgs } from '@remix-run/node';
import { useLoaderData, Link, useNavigate } from '@remix-run/react';
import { Building, Database, Cog, Link as LinkIcon, LayoutGrid, Trash2, Info, FileText, Blocks, Terminal, Settings, ArrowLeft, Activity } from 'lucide-react';
import { getService } from '~/actions/service.server';
import ServiceDeleteModal from '~/components/services/ServiceDeleteModal';
import InformationTab from '~/components/services/InformationTab';
import IntegrationTab from '~/components/services/IntegrationTab';
import EnvironmentVariablesTab from '~/components/services/EnvironmentVariablesTab';
import LogsTab from '~/components/services/LogsTab';
import DeploymentTab from '~/components/services/DeploymentTab';
import ConfigurationTab from '~/components/services/ConfigurationTab';

export async function loader({ params, request }: LoaderFunctionArgs) {
  try {
    const serviceId = params.id;
    if (!serviceId) {
      throw new Response('Service ID is required', { status: 400 });
    }
    
    const service = await getService(serviceId, request);
    return json({ service });
  } catch (error) {
    console.error('Error loading service:', error);
    throw new Response('Service not found', { status: 404 });
  }
}

export default function ServiceDetail() {
  const { service } = useLoaderData<typeof loader>();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const navigate = useNavigate();
  
  // State to track the active tab - start with 'information' as default
  const [activeTab, setActiveTab] = useState('information');
  
  // Handle tab change by updating both state and URL hash
  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId);
    if (typeof window !== 'undefined') {
      window.location.hash = tabId;
    }
  };
  
  // Get tab from hash helper function (client-side only)
  const getTabFromHash = () => {
    if (typeof window === 'undefined') return 'information';
    // Get the hash from URL, removing the '#' character
    const hash = window.location.hash.substring(1);
    // Return the hash if it's a valid tab, otherwise default to 'information'
    return hash && allTabs.some(tab => tab.id === hash) ? hash : 'information';
  };
  
  // Initialize tab from hash and listen for hash changes (client-side only)
  useEffect(() => {
    if (typeof window === 'undefined') return;
    
    // Set initial tab based on URL hash when component mounts on client
    const initialTab = getTabFromHash();
    if (initialTab !== activeTab) {
      setActiveTab(initialTab);
    }
    
    // Listen for hash changes
    const handleHashChange = () => {
      const newTab = getTabFromHash();
      if (newTab !== activeTab) {
        setActiveTab(newTab);
      }
    };
    
    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, [activeTab]);
  
  // Ensure we don't show hidden tabs for managed services
  useEffect(() => {
    // If service is managed and active tab is integration, switch to information tab
    if (service.type === 'managed' && activeTab === 'integration') {
      handleTabChange('information');
    }
  }, [service.type, activeTab]);
  
  // Set URL hash on initial load if it doesn't exist (client-side only)
  useEffect(() => {
    if (typeof window === 'undefined') return;
    
    if (!window.location.hash) {
      window.location.hash = activeTab;
    }
  }, [activeTab]);

  // Define all possible tabs
  const allTabs = [
    { id: 'information', name: 'Information', icon: <Info className="w-4 h-4" /> },
    { id: 'deployment', name: 'Deployment', icon: <Building className="w-4 h-4" /> , hideForManaged: true },
    { id: 'logs', name: 'Logs', icon: <FileText className="w-4 h-4" /> , hideForManaged: true },
    { id: 'environment', name: 'Environment', icon: <Database className="w-4 h-4" /> },
    { id: 'integration', name: 'Integration', icon: <LinkIcon className="w-4 h-4" />, hideForManaged: true },
    { id: 'configuration', name: 'Configuration', icon: <Cog className="w-4 h-4" /> },
  ];
  
  // Filter tabs based on service type
  const tabs = allTabs.filter(tab => {
    if (service.type === 'managed' && tab.hideForManaged) {
      return false;
    }
    return true;
  });

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-8">
          {/* Header Section */}
          <div className="space-y-4">
            <Link 
              to={`/projects/${service.projectId}`}
              className="inline-flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors group"
            >
              <ArrowLeft size={16} className="mr-2 transition-transform group-hover:-translate-x-1" />
              Back to Project
            </Link>
            
            <div className="text-left justify-between flex items-center">
              <div className='space-y-3'>
              <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl flex items-center justify-center">
                <LayoutGrid className="w-8 h-8 mr-3" />
                {service.name}
              </h1>
              
              {/* Service Badges */}
              <div className="flex items-center justify-center space-x-3">
                <span className={`inline-flex items-center px-3 py-1 rounded-lg text-xs font-semibold ${
                  service.type === 'git' 
                    ? 'bg-blue-100/80 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400 border border-blue-200/50 dark:border-blue-700/50' 
                    : 'bg-purple-100/80 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400 border border-purple-200/50 dark:border-purple-700/50'
                } backdrop-blur-sm`}>
                  {service.type === 'git' ? 'Git Repository' : `Managed ${service.managedType}`}
                </span>
                
                <span className={`inline-flex items-center px-3 py-1 rounded-lg text-xs font-semibold backdrop-blur-sm border ${
                  !service.status || service.status === 'pending' 
                    ? 'bg-yellow-100/80 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 border-yellow-200/50 dark:border-yellow-700/50' 
                    : service.status === 'running' || service.status === 'active'
                      ? 'bg-green-100/80 text-green-800 dark:bg-green-900/30 dark:text-green-400 border-green-200/50 dark:border-green-700/50'
                      : 'bg-red-100/80 text-red-800 dark:bg-red-900/30 dark:text-red-400 border-red-200/50 dark:border-red-700/50'
                }`}>
                  <span className={`w-1.5 h-1.5 rounded-full mr-2 ${
                    !service.status || service.status === 'pending' ? 'bg-yellow-500' :
                    service.status === 'running' || service.status === 'active' ? 'bg-green-500' : 'bg-red-500'
                  }`} />
                  {service.status ? service.status.charAt(0).toUpperCase() + service.status.slice(1) : 'Pending'}
                </span>
              </div>
            </div>
             {/* Delete Button */}
             <div className="flex justify-end">
              <button
                onClick={() => setIsDeleteModalOpen(true)}
                className="inline-flex items-center px-4 py-2 bg-red-100/80 dark:bg-red-900/30 text-red-700 dark:text-red-300 font-medium rounded-xl border border-red-200/50 dark:border-red-700/50 hover:bg-red-200/80 dark:hover:bg-red-900/50 transition-all duration-200 backdrop-blur-sm"
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Delete Service
              </button>
            </div>
              </div>
           
          </div>
          
          {/* Main Content */}
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl overflow-hidden">
            {/* Tabs */}
            <div className="border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
              <nav className="flex" aria-label="Tabs">
                {tabs.map(tab => (
                  <button
                    key={tab.id}
                    onClick={() => handleTabChange(tab.id)}
                    className={`flex-1 py-4 px-6 text-center border-b-2 font-medium text-sm transition-all duration-200 ${
                      activeTab === tab.id
                        ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400 bg-indigo-50/50 dark:bg-indigo-900/20'
                        : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 hover:bg-gray-50/50 dark:hover:bg-gray-700/30'
                    }`}
                  >
                    <div className="flex flex-col items-center space-y-1">
                      {tab.icon}
                      <span className="text-xs">{tab.name}</span>
                    </div>
                  </button>
                ))}
              </nav>
            </div>
            
            {/* Tab content */}
            <div className="min-h-96">
              {activeTab === 'information' && <InformationTab service={service} />}
              {activeTab === 'integration' && <IntegrationTab service={service} />}
              {activeTab === 'environment' && <EnvironmentVariablesTab service={service} />}
              {activeTab === 'logs' && <LogsTab service={service} />}
              {activeTab === 'deployment' && <DeploymentTab service={service} />}
              {activeTab === 'configuration' && (
                <ConfigurationTab 
                  service={service} 
                  onServiceUpdated={() => {
                    // Reload service data when configuration is updated
                    window.location.reload();
                  }}
                />
              )}
            </div>
          </div>
        </div>
      </div>
      
      {/* Delete Modal */}
      <ServiceDeleteModal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        service={service}
        onSuccess={() => {
          setIsDeleteModalOpen(false);
          navigate(`/projects/${service.projectId}`);
        }}
      />
    </div>
  );
}