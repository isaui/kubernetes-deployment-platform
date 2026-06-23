import React, { useState } from 'react';
import type { PodStats } from '~/types/stats';
import { formatBytesConsistently } from '~/utils/formatters';

interface PodsListProps {
  pods: PodStats[];
}

export function PodsList({ pods }: PodsListProps) {
  const [searchQuery, setSearchQuery] = useState('');

  console.log(pods)
  
  // Filter pods based on search query
  const filteredPods = pods.filter(pod => 
    pod.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    pod.namespace.toLowerCase().includes(searchQuery.toLowerCase()) ||
    pod.status.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group pods by status
  const statusCounts = pods.reduce((acc, pod) => {
    acc[pod.status] = (acc[pod.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const getStatusColor = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'bg-green-100 text-green-800';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'succeeded':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      {/* Search bar */}
      <div className="bg-white shadow rounded-lg p-4">
        <label htmlFor="pod-search" className="sr-only">Search pods</label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
            </svg>
          </div>
          <input
            id="pod-search"
            name="pod-search"
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Search pods by name, namespace or status..."
            type="search"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Pod status summary */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-gray-500">Total Pods</h3>
              <p className="text-lg font-semibold text-gray-900">{pods.length}</p>
            </div>
          </div>
        </div>
        {Object.entries(statusCounts).map(([status, count]) => (
          <div key={status} className="bg-white shadow rounded-lg p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className={`h-6 w-6 rounded-full ${getStatusColor(status)} flex items-center justify-center`}>
                  <span className="text-xs font-medium">{count}</span>
                </div>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-gray-500">{status}</h3>
                <p className="text-lg font-semibold text-gray-900">{count}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Pods table */}
      <div className="bg-white shadow overflow-hidden rounded-lg">
        <div className="px-4 py-5 border-b border-gray-200 sm:px-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Pods ({filteredPods.length})
          </h3>
        </div>
        <div className="overflow-x-auto">
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
                  CPU
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  CPU Request
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  CPU Limit
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Memory
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Memory Request
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Memory Limit
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredPods.map((pod) => (
                <tr key={`${pod.namespace}-${pod.name}`}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {pod.name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {pod.namespace}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(pod.status)}`}>
                      {pod.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.cpu === 'object' ? 
                      `${pod.cpu.usage || '0'} (${pod.cpu.percentage ? `${pod.cpu.percentage.toFixed(1)}%` : '0%'})` : 
                      pod.cpu}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.cpu === 'object' && pod.cpu.request ? pod.cpu.request : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.cpu === 'object' && pod.cpu.limit ? pod.cpu.limit : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.memory === 'object' ? 
                      `${formatBytesConsistently(pod.memory.usage) || '0 B'} (${pod.memory.percentage ? `${pod.memory.percentage.toFixed(1)}%` : '0%'})` : 
                      pod.memory}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.memory === 'object' && pod.memory.request ? formatBytesConsistently(pod.memory.request) : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {typeof pod.memory === 'object' && pod.memory.limit ? formatBytesConsistently(pod.memory.limit) : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(pod.created).toLocaleDateString()}
                  </td>
                </tr>
              ))}
              {filteredPods.length === 0 && (
                <tr>
                  <td colSpan={10} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center">
                    No pods found matching your search.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
