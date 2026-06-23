import React, { useState } from 'react';
import { 
  ResponsiveContainer, PieChart, Pie, Cell, Tooltip, Label
} from 'recharts';
import type { NodeStats } from '~/types/stats';
import { formatBytesConsistently } from '~/utils/formatters';

interface NodesOverviewProps {
  nodes: NodeStats[];
}

export default function NodesOverview({ nodes }: NodesOverviewProps) {
  const [searchQuery, setSearchQuery] = useState('');
  
  // Filter nodes based on search query
  const filteredNodes = nodes.filter(node => 
    node.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Calculate average CPU, memory, and storage usage percentages
  const calculateAveragePercentage = (nodes: NodeStats[], key: string): number => {
    if (!nodes.length) return 0;
    
    const total = nodes.reduce((sum, node:any) => {
      return sum + (key === 'storage' ? 
        (node.storage?.percentage || 0) : 
        node[key].percentage);
    }, 0);
    
    return Math.round(total / nodes.length);
  };

  // Function to format memory values consistently
  const formatMemoryConsistently = (usage: string, capacity: string): string => {
    // Convert both to human-readable format with the same unit
    return `${formatBytesConsistently(usage)} / ${formatBytesConsistently(capacity)}`;
  };

  const avgCpuUsage = calculateAveragePercentage(nodes, 'cpu');
  const avgMemoryUsage = calculateAveragePercentage(nodes, 'memory');
  const avgStorageUsage = calculateAveragePercentage(nodes, 'storage');
  
  // Prepare data for gauge charts
  const createGaugeData = (usedPercentage:any) => [
    { name: 'Used', value: usedPercentage },
    { name: 'Available', value: Math.max(0, 100 - usedPercentage) }
  ];
  
  const cpuGaugeData = createGaugeData(avgCpuUsage);
  const memoryGaugeData = createGaugeData(avgMemoryUsage);
  const storageGaugeData = createGaugeData(avgStorageUsage);

  return (
    <div className="space-y-8">
      {/* Search bar */}
      <div className="bg-white shadow rounded-lg p-4">
        <label htmlFor="search" className="sr-only">Search nodes</label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
            </svg>
          </div>
          <input
            id="search"
            name="search"
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Search nodes..."
            type="search"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Node count summary */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-bold mb-4">Nodes Summary</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Total Nodes</h3>
            <p className="mt-1 text-lg font-semibold">{nodes.length}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Ready Nodes</h3>
            <p className="mt-1 text-lg font-semibold">
              {nodes.filter(node => node.status === "Ready").length}
            </p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Not Ready Nodes</h3>
            <p className="mt-1 text-lg font-semibold">
              {nodes.filter(node => node.status !== "Ready").length}
            </p>
          </div>
        </div>
      </div>

      {/* Resource usage charts */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        {/* CPU Gauge */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium mb-2 text-center">CPU Usage</h3>
          <div className="h-64 flex flex-col items-center justify-center">
            <ResponsiveContainer width="100%" height="70%">
              <PieChart>
                <Pie
                  data={[
                    { name: 'Used', value: avgCpuUsage || 0, fill: '#8884d8' },
                    { name: 'Free', value: 100 - (avgCpuUsage || 0), fill: '#f3f4f6' }
                  ]}
                  cx="50%"
                  cy="50%"
                  startAngle={180}
                  endAngle={0}
                  innerRadius="60%"
                  outerRadius="80%"
                  paddingAngle={0}
                  dataKey="value"
                >
                  {/* Add label inside gauge */}
                  <Label
                    value={`${avgCpuUsage ? avgCpuUsage.toFixed(1) : 0}%`}
                    position="center"
                    className="text-xl font-medium"
                    fill="#333333"
                  />
                </Pie>
                <Tooltip 
                  formatter={(value: number) => [`${value.toFixed(1)}%`, 'Usage']}
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      return (
                        <div className="bg-white p-2 border rounded shadow text-xs">
                          <p>Average CPU Usage: {Number(payload[0].value).toFixed(1)}%</p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div className="mt-2 text-2xl font-semibold text-gray-800">
              {avgCpuUsage}%
            </div>
            <div className="text-sm text-gray-500">Average CPU Usage</div>
          </div>
        </div>

        {/* Memory Gauge */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium mb-2 text-center">Memory Usage</h3>
          <div className="h-64 flex flex-col items-center justify-center">
            <ResponsiveContainer width="100%" height="70%">
              <PieChart>
                <Pie
                  data={[
                    { name: 'Used', value: avgMemoryUsage || 0, fill: '#82ca9d' },
                    { name: 'Free', value: 100 - (avgMemoryUsage || 0), fill: '#f3f4f6' }
                  ]}
                  cx="50%"
                  cy="50%"
                  startAngle={180}
                  endAngle={0}
                  innerRadius="60%"
                  outerRadius="80%"
                  paddingAngle={0}
                  dataKey="value"
                >
                  {/* Add label inside gauge */}
                  <Label
                    value={`${avgMemoryUsage ? avgMemoryUsage.toFixed(1) : 0}%`}
                    position="center"
                    className="text-xl font-medium"
                    fill="#333333"
                  />
                </Pie>
                <Tooltip 
                  formatter={(value: number) => [`${value.toFixed(1)}%`, 'Usage']}
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      return (
                        <div className="bg-white p-2 border rounded shadow text-xs">
                          <p>Average Memory Usage: {Number(payload[0].value).toFixed(1)}%</p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div className="mt-2 text-2xl font-semibold text-gray-800">
              {avgMemoryUsage}%
            </div>
            <div className="text-sm text-gray-500">Average Memory Usage</div>
          </div>
        </div>
        
        {/* Storage Gauge */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium mb-2 text-center">Storage Usage</h3>
          <div className="h-64 flex flex-col items-center justify-center">
            <ResponsiveContainer width="100%" height="70%">
              <PieChart>
                <Pie
                  data={[
                    { name: 'Used', value: avgStorageUsage || 0, fill: '#ff8042' },
                    { name: 'Free', value: 100 - (avgStorageUsage || 0), fill: '#f3f4f6' }
                  ]}
                  cx="50%"
                  cy="50%"
                  startAngle={180}
                  endAngle={0}
                  innerRadius="60%"
                  outerRadius="80%"
                  paddingAngle={0}
                  dataKey="value"
                >
                  {/* Add label inside gauge */}
                  <Label
                    value={`${avgStorageUsage ? avgStorageUsage.toFixed(1) : 0}%`}
                    position="center"
                    className="text-xl font-medium"
                    fill="#333333"
                  />
                </Pie>
                <Tooltip 
                  formatter={(value: number) => [`${value.toFixed(1)}%`, 'Usage']}
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      return (
                        <div className="bg-white p-2 border rounded shadow text-xs">
                          <p>Average Storage Usage: {Number(payload[0].value).toFixed(1)}%</p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div className="mt-2 text-2xl font-semibold text-gray-800">
              {avgStorageUsage}%
            </div>
            <div className="text-sm text-gray-500">Average Storage Usage</div>
          </div>
        </div>
      </section>

      {/* Nodes Table */}
      <section className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 py-5 border-b border-gray-200 sm:px-6">
          <h2 className="text-lg font-medium text-gray-900">Nodes List ({filteredNodes.length})</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Roles
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  CPU
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Memory
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Storage
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredNodes.map((node) => (
                <tr key={node.name}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {node.name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      node.status === 'Ready' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {node.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {node.roles.join(', ')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <div className="flex items-center">
                      <div className="w-full bg-gray-200 h-2 rounded-full overflow-hidden">
                        <div 
                          className="bg-blue-500 h-full rounded-full" 
                          style={{ width: `${Math.min(100, node.cpu.percentage)}%` }}
                        ></div>
                      </div>
                      <span className="text-xs font-medium">
                        {node.cpu.usage} / {node.cpu.capacity} cores ({node.cpu.percentage ? `${node.cpu.percentage.toFixed(1)}%` : '0%'})
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <div className="flex items-center">
                      <div className="w-full bg-gray-200 h-2 rounded-full overflow-hidden">
                        <div 
                          className="bg-green-500 h-full rounded-full" 
                          style={{ width: `${Math.min(100, node.memory.percentage)}%` }}
                        ></div>
                      </div>
                      <span className="text-xs font-medium">
                        {formatMemoryConsistently(node.memory.usage, node.memory.capacity)} ({node.memory.percentage ? `${node.memory.percentage.toFixed(1)}%` : '0%'})
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <div className="flex items-center">
                      <div className="w-full bg-gray-200 h-2 rounded-full overflow-hidden">
                        <div 
                          className="bg-yellow-500 h-full rounded-full" 
                          style={{ width: `${Math.min(100, node.storage ? 
                            (node.storage.percentage !== undefined ? 
                              node.storage.percentage : 
                              node.storage.capacity && node.storage.usage ? 
                                (parseFloat(node.storage.usage.toString().replace(/[^\d.]/g, '')) / 
                                 parseFloat(node.storage.capacity.toString().replace(/[^\d.]/g, ''))) * 100 : 
                                0) : 0)}%` }}
                        ></div>
                      </div>
                      <span className="text-xs font-medium">
                        {node.storage ? 
                          `${formatBytesConsistently(node.storage.usage)} / ${formatBytesConsistently(node.storage.capacity)} (${node.storage.percentage !== undefined ? 
                            `${node.storage.percentage.toFixed(1)}%` : 
                            node.storage.capacity && node.storage.usage ? 
                              `${((parseFloat(node.storage.usage.toString().replace(/[^\d.]/g, '')) / 
                                parseFloat(node.storage.capacity.toString().replace(/[^\d.]/g, ''))) * 100).toFixed(1)}%` : 
                              '0%'})` : 
                          '0 B / 0 B (0%)'}  
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(node.created).toLocaleDateString()}
                  </td>
                </tr>
              ))}
              {filteredNodes.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center">
                    No nodes found matching your search.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
