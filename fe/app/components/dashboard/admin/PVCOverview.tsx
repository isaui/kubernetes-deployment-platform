import React, { useState } from 'react';
import { PVCStats } from '~/types/stats';
import { formatDate } from '~/utils/formatters';

interface PVCOverviewProps {
  pvcs: PVCStats[];
}

export default function PVCOverview({ pvcs }: PVCOverviewProps) {
  const [selectedNamespace, setSelectedNamespace] = useState<string>('all');
  
  // Get unique namespaces for filter dropdown
  const namespaces = ['all', ...Array.from(new Set(pvcs.map(pvc => pvc.namespace)))];
  
  // Filter PVCs by selected namespace
  const filteredPVCs = selectedNamespace === 'all' 
    ? pvcs 
    : pvcs.filter(pvc => pvc.namespace === selectedNamespace);

  return (
    <div>
      <div className="mb-6">
        <div className="flex justify-between items-center">
          <h2 className="text-xl font-semibold text-gray-900">Persistent Volume Claims</h2>
          <div className="flex items-center space-x-2">
            <label htmlFor="namespaceFilter" className="text-sm font-medium text-gray-700">
              Namespace:
            </label>
            <select
              id="namespaceFilter"
              value={selectedNamespace}
              onChange={(e) => setSelectedNamespace(e.target.value)}
              className="rounded-md border-gray-300 py-2 pl-3 pr-10 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
            >
              {namespaces.map((namespace) => (
                <option key={namespace} value={namespace}>
                  {namespace}
                </option>
              ))}
            </select>
          </div>
        </div>
        <p className="mt-1 text-sm text-gray-500">
          Total: {filteredPVCs.length} PVCs {selectedNamespace !== 'all' && `in namespace ${selectedNamespace}`}
        </p>
      </div>

      <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Namespace
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Storage
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Access Modes
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Storage Class
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Created
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredPVCs.length > 0 ? (
              filteredPVCs.map((pvc) => (
                <tr key={`${pvc.namespace}-${pvc.name}`}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {pvc.name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {pvc.namespace}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                        ${pvc.status === 'Bound' ? 'bg-green-100 text-green-800' : 
                          pvc.status === 'Pending' ? 'bg-yellow-100 text-yellow-800' :
                          'bg-red-100 text-red-800'}`}
                    >
                      {pvc.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {pvc.storageCapacity}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {pvc.accessModes.join(', ')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {pvc.storageClassName || 'default'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(pvc.created)}
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={7} className="px-6 py-4 text-center text-sm text-gray-500">
                  No PVCs found {selectedNamespace !== 'all' && `in namespace ${selectedNamespace}`}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
