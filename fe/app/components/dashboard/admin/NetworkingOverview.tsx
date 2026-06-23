import { useState } from 'react';
import type { ServiceStats, IngressStats } from '~/types/stats';
import { ServicesList } from './networking/ServicesList';
import { IngressList } from './networking/IngressList';


interface NetworkingOverviewProps {
  services: ServiceStats[];
  ingresses: IngressStats[];
}

export default function NetworkingOverview({ services, ingresses }: NetworkingOverviewProps) {
  const [activeNetworkingView, setActiveNetworkingView] = useState<'services' | 'ingresses'>('services');

  return (
    <div className="space-y-6">
      {/* Networking type selector */}
      <div className="bg-white shadow rounded-lg p-4">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-xl font-semibold text-gray-900">Networking</h1>
            <p className="mt-2 text-sm text-gray-700">
              View and manage services and ingress resources across your cluster
            </p>
          </div>
          <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
            <div className="flex space-x-3">
              <button
                type="button"
                onClick={() => setActiveNetworkingView('services')}
                className={`inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md 
                  ${activeNetworkingView === 'services' 
                    ? 'bg-indigo-600 text-white hover:bg-indigo-700' 
                    : 'bg-white text-gray-700 hover:bg-gray-50 border-gray-300'
                  }`}
              >
                Services
              </button>
              <button
                type="button"
                onClick={() => setActiveNetworkingView('ingresses')}
                className={`inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md 
                  ${activeNetworkingView === 'ingresses' 
                    ? 'bg-indigo-600 text-white hover:bg-indigo-700' 
                    : 'bg-white text-gray-700 hover:bg-gray-50 border-gray-300'
                  }`}
              >
                Ingresses
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Networking summary */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-medium mb-4">Networking Summary</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Total Services</h3>
            <p className="mt-1 text-lg font-semibold">{services.length}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Total Ingresses</h3>
            <p className="mt-1 text-lg font-semibold">{ingresses.length}</p>
          </div>
        </div>
      </div>

      {/* Display appropriate component based on active view */}
      {activeNetworkingView === 'services' ? (
        <ServicesList services={services} />
      ) : (
        <IngressList ingresses={ingresses} />
      )}
    </div>
  );
}
