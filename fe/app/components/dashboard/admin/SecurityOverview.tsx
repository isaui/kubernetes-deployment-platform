import React from 'react';
import type { CertificateStats } from '~/types/stats';
import { CertificatesList } from './security/CertificatesList';

interface SecurityOverviewProps {
  certificates: CertificateStats[];
}

export default function SecurityOverview({ certificates }: SecurityOverviewProps) {
  // Count valid and expiring certificates
  const validCerts = certificates.filter(cert => 
    !cert.isExpired
  );
  
  const expiringCerts = certificates.filter(cert => {
    return !cert.isExpired && cert.daysUntilExpiry <= 30;
  });

  const expiredCerts = certificates.filter(cert => 
    cert.isExpired
  );

  return (
    <div className="space-y-6">
      {/* Security overview */}
      <div className="bg-white shadow rounded-lg p-4">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-xl font-semibold text-gray-900">Security</h1>
            <p className="mt-2 text-sm text-gray-700">
              Monitor and manage certificates and security resources
            </p>
          </div>
        </div>
      </div>

      {/* Certificate summary stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-green-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-gray-500">Valid Certificates</h3>
              <p className="text-lg font-semibold text-gray-900">{validCerts.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-yellow-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-gray-500">Expiring Soon (30d)</h3>
              <p className="text-lg font-semibold text-gray-900">{expiringCerts.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-red-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-gray-500">Expired Certificates</h3>
              <p className="text-lg font-semibold text-gray-900">{expiredCerts.length}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Certificates list */}
      <CertificatesList certificates={certificates} />
    </div>
  );
}
