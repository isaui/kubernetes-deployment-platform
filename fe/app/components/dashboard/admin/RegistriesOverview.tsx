import React, { useState } from 'react';
import type { Registry } from '~/types/registry';
import { fetchRegistries, deleteRegistry } from '~/actions/registry';

interface RegistriesOverviewProps {
  initialRegistries?: Registry[];
}

export default function RegistriesOverview({ initialRegistries = [] }: RegistriesOverviewProps) {
  const [registries, setRegistries] = useState<Registry[]>(initialRegistries);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState<number>(1);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false);
  const [registryToDelete, setRegistryToDelete] = useState<Registry | null>(null);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);

  const loadRegistries = async (pageNum = 1) => {
    setIsLoading(true);
    setError(null);
    try {
      console.log('Loading registries...');
      const response = await fetchRegistries({ 
        page: pageNum, 
        pageSize: 10,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      });
      console.log('Registries response:', response.registries);
      setRegistries(response.registries);
      setTotalPages(response.totalPages);
      setPage(pageNum);
    } catch (err: any) {
      setError(err.message || 'Failed to load registries');
    } finally {
      setIsLoading(false);
    }
  };

  // Load registries on initial render if none provided
  React.useEffect(() => {
    if (!initialRegistries || initialRegistries.length === 0) {
      loadRegistries();
    }
  }, [initialRegistries]);

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'ready':
        return 'bg-green-100 text-green-800';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'building':
        return 'bg-blue-100 text-blue-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="mt-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold text-gray-800">Container Registries</h2>
        <a
          href="/admin/registries/new"
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
        >
          Add Registry
        </a>
      </div>

      {error && (
        <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-4">
          <div className="flex">
            <div className="ml-3">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          </div>
        </div>
      )}

      <div className="bg-white shadow overflow-hidden rounded-lg overflow-x-auto">
        {isLoading ? (
          <div className="p-6 text-center text-gray-500">Loading...</div>
        ) : registries.length === 0 ? (
          <div className="p-6 text-center text-gray-500">No registries found</div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  URL
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Default
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {registries.map((registry) => (
                <tr key={registry.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => window.location.href = `/admin/registries/${registry.id}`}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {registry.name}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 max-w-[200px]">
                    <div className="break-all">{registry.url}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusBadgeClass(registry.status)}`}>
                      {registry.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {registry.isDefault ? 'Yes' : 'No'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(registry.createdAt).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <a
                      href={`/admin/registries/${registry.id}`}
                      className="text-indigo-600 hover:text-indigo-900 mr-4"
                    >
                      View
                    </a>
                    <a
                      href={`/admin/registries/${registry.id}/edit`}
                      className="text-indigo-600 hover:text-indigo-900 mr-4"
                    >
                      Edit
                    </a>
                    <button
                      className="text-red-600 hover:text-red-900"
                      onClick={(e) => {
                        e.stopPropagation(); // Prevent row click navigation
                        setRegistryToDelete(registry);
                        setShowDeleteModal(true);
                      }}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex justify-between items-center mt-4">
          <button
            onClick={() => loadRegistries(page - 1)}
            disabled={page === 1}
            className={`px-4 py-2 border rounded-md ${
              page === 1 ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Previous
          </button>
          <div className="text-sm text-gray-700">
            Page {page} of {totalPages}
          </div>
          <button
            onClick={() => loadRegistries(page + 1)}
            disabled={page === totalPages}
            className={`px-4 py-2 border rounded-md ${
              page === totalPages ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Next
          </button>
        </div>
      )}
      
      {/* Delete Modal */}
      {showDeleteModal && registryToDelete && (
        <div className="fixed z-20 inset-0 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            {/* Background overlay */}
            <div 
              className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              aria-hidden="true"
              onClick={() => !isDeleting && setShowDeleteModal(false)}
            ></div>

            {/* Modal panel */}
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="sm:flex sm:items-start">
                  <div className="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                    <svg className="h-6 w-6 text-red-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  </div>
                  <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <h3 className="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                      Delete Registry
                    </h3>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">
                        Are you sure you want to delete the registry <span className="font-medium">{registryToDelete.name}</span>? This action cannot be undone.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  disabled={isDeleting}
                  className={`w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 ${isDeleting ? 'bg-red-300' : 'bg-red-600 hover:bg-red-700'} text-base font-medium text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm`}
                  onClick={async () => {
                    try {
                      setIsDeleting(true);
                      setError(null);
                      
                      // Call the API to delete the registry
                      await deleteRegistry(registryToDelete.id);
                      
                      // Fetch fresh data from server after successful delete
                      await loadRegistries(page);
                      
                      // Close the modal
                      setShowDeleteModal(false);
                      setRegistryToDelete(null);
                    } catch (err: any) {
                      setError(`Failed to delete registry: ${err.message}`);
                    } finally {
                      setIsDeleting(false);
                    }
                  }}
                >
                  {isDeleting ? 'Deleting...' : 'Delete'}
                </button>
                <button
                  type="button"
                  disabled={isDeleting}
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => {
                    if (!isDeleting) {
                      setShowDeleteModal(false);
                      setRegistryToDelete(null);
                    }
                  }}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
