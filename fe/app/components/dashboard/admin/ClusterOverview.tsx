import React from 'react';
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
  PieChart, Pie, Cell
} from 'recharts';
import type { ClusterInfo } from '~/types/stats';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

interface ClusterOverviewProps {
  clusterInfo: ClusterInfo;
}

export default function ClusterOverview({ clusterInfo }: ClusterOverviewProps) {
  // Prepare data for resource distribution chart
  const distributionData = [
    { name: 'Nodes', value: clusterInfo.stats.nodeCount },
    { name: 'Namespaces', value: clusterInfo.stats.namespaceCount },
    { name: 'Pods', value: clusterInfo.stats.podCount }
  ];

  return (
    <div className="space-y-8">
      {/* Cluster Information */}
      <section className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-bold mb-4">Cluster Information</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Kubernetes Version</h3>
            <p className="mt-1 text-lg font-semibold">{clusterInfo.version.gitVersion}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Platform</h3>
            <p className="mt-1 text-lg font-semibold">{clusterInfo.version.platform}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Go Version</h3>
            <p className="mt-1 text-lg font-semibold">{clusterInfo.version.goVersion}</p>
          </div>
          <div className="bg-gray-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-gray-500">Build Date</h3>
            <p className="mt-1 text-lg font-semibold">{clusterInfo.version.buildDate}</p>
          </div>
        </div>
      </section>

      {/* Cluster Stats */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-base font-medium text-gray-500">Nodes</h3>
          <p className="mt-1 text-3xl font-semibold">{clusterInfo.stats.nodeCount}</p>
          <div className="mt-2 flex items-center text-sm text-gray-500">
            <span className="truncate">Total nodes in the cluster</span>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-base font-medium text-gray-500">Namespaces</h3>
          <p className="mt-1 text-3xl font-semibold">{clusterInfo.stats.namespaceCount}</p>
          <div className="mt-2 flex items-center text-sm text-gray-500">
            <span className="truncate">Total namespaces in the cluster</span>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-base font-medium text-gray-500">Pods</h3>
          <p className="mt-1 text-3xl font-semibold">{clusterInfo.stats.podCount}</p>
          <div className="mt-2 flex items-center text-sm text-gray-500">
            <span className="truncate">Total pods in the cluster</span>
          </div>
        </div>
      </section>

      {/* Charts */}
      <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium mb-4">Resource Distribution</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={distributionData}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="value" fill="#8884d8" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium mb-4">Resource Composition</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={distributionData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                  label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                >
                  {distributionData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </section>
    </div>
  );
}
