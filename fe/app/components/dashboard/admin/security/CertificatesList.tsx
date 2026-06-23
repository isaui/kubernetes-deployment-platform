import React, { useState } from 'react';
import type { CertificateStats } from '~/types/stats';

interface CertificatesListProps {
  certificates: CertificateStats[];
}

export function CertificatesList({ certificates }: CertificatesListProps) {
  const [searchQuery, setSearchQuery] = useState('');
  
  // Filter certificates based on search query
  const filteredCertificates = certificates.filter(cert => 
    cert.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    cert.namespace.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (cert.domains && cert.domains.some(dns => dns?.toLowerCase().includes(searchQuery.toLowerCase())))
  );

  const getCertificateStatus = (cert: CertificateStats): {
    status: 'valid' | 'expiring' | 'expired';
    statusDisplay: string;
    statusClass: string;
  } => {
    // Use the built-in isExpired and daysUntilExpiry properties
    if (cert.isExpired) {
      return {
        status: 'expired',
        statusDisplay: 'Expired',
        statusClass: 'bg-red-100 text-red-800'
      };
    } else if (cert.daysUntilExpiry <= 30) {
      return {
        status: 'expiring',
        statusDisplay: 'Expiring Soon',
        statusClass: 'bg-yellow-100 text-yellow-800'
      };
    } else {
      return {
        status: 'valid',
        statusDisplay: 'Valid',
        statusClass: 'bg-green-100 text-green-800'
      };
    }
  };

  // We don't need this function anymore as we use cert.daysUntilExpiry directly;

  return (
    <div className="space-y-6">
      {/* Search bar */}
      <div className="bg-white shadow rounded-lg p-4">
        <label htmlFor="certificate-search" className="sr-only">Search certificates</label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
            </svg>
          </div>
          <input
            id="certificate-search"
            name="certificate-search"
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Search certificates by name, namespace or DNS names..."
            type="search"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Certificates table */}
      <div className="bg-white shadow overflow-hidden rounded-lg">
        <div className="px-4 py-5 border-b border-gray-200 sm:px-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Certificates ({filteredCertificates.length})
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
                  DNS Names
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Issuer
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Expiration
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredCertificates.length > 0 ? (
                filteredCertificates.map((cert) => {
                  const { statusDisplay, statusClass } = getCertificateStatus(cert);
                  const daysRemaining = cert.daysUntilExpiry;
                  
                  return (
                    <tr key={`${cert.namespace}-${cert.name}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {cert.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {cert.namespace}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        <div className="flex flex-col gap-1">
                          {cert.domains && cert.domains.length > 0 ? (
                            cert.domains.map((dns, index) => (
                              <span 
                                key={index}
                                className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800"
                              >
                                {dns}
                              </span>
                            ))
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {cert.issuer}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusClass}`}>
                          {statusDisplay}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <div className="flex flex-col">
                          <span className="text-sm">
                            {cert.isExpired ? 'Expired' : `Expires in ${cert.daysUntilExpiry} days`}
                          </span>
                          <span className={`text-xs ${
                            cert.isExpired 
                              ? 'text-red-600'
                              : daysRemaining <= 30
                                ? 'text-yellow-600'
                                : 'text-green-600'
                          }`}>
                            {cert.isExpired 
                              ? 'Certificate expired' 
                              : `${daysRemaining} days remaining`}
                          </span>
                        </div>
                      </td>
                    </tr>
                  );
                })
              ) : (
                <tr>
                  <td colSpan={6} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center">
                    No certificates found matching your search.
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
