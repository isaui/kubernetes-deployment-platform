import { json, redirect, type LoaderFunctionArgs, type ActionFunctionArgs } from '@remix-run/node';
import { useLoaderData, useNavigate, Form, useSubmit } from '@remix-run/react';
import { useState } from 'react';
import { getProjects, deleteProject } from '~/actions/project.server';
import DeleteConfirmationModal from '~/components/projects/ProjectDeletionModal';
import ProjectList from '~/components/projects/ProjectList';
import type { ProjectFilter, ProjectListResponse, Project } from '~/types/project';

// Define loader data type to include optional error
interface LoaderData {
  projectsData: ProjectListResponse;
  filter: ProjectFilter;
  error?: string;
}

// Also protect the action function
export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const projectId = formData.get('projectId') as string;
  const actionType = formData.get('_action') as string;
  
  if (actionType === 'delete' && projectId) {
    try {
      await deleteProject(projectId, request);
      // Redirect ke halaman yang sama untuk me-reload data
      return redirect('/projects');
    } catch (error) {
      return json(
        { error: 'Failed to delete project. Please try again.' },
        { status: 500 }
      );
    }
  }
  
  return json({ error: 'Invalid action' }, { status: 400 });
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = parseInt(url.searchParams.get('page') || '1', 10);
  const pageSize = parseInt(url.searchParams.get('pageSize') || '9', 10);
  const search = url.searchParams.get('search') || '';
  const sortBy = url.searchParams.get('sortBy') || 'createdAt';
  const sortOrder = url.searchParams.get('sortOrder') || 'desc';

  const filter: ProjectFilter = {
    page,
    pageSize,
    search,
    sortBy,
    sortOrder
  };

  try {
    const projectsData = await getProjects(filter, request);
    // Pastikan projectsData.projects adalah array meskipun kosong
    return json<LoaderData>({ 
      projectsData: {
        ...projectsData,
        projects: projectsData.projects || [] 
      }, 
      filter 
    });
  } catch (error) {
    console.error('Error loading projects:', error);
    return json<LoaderData>({ 
      projectsData: { 
        projects: [], 
        totalCount: 0, 
        page: 1, 
        pageSize: 9, 
        totalPages: 0 
      }, 
      filter,
      error: 'Failed to load projects'
    });
  }
}

export default function ProjectsIndex() {
  const { projectsData, filter, error } = useLoaderData<LoaderData>();
  const navigate = useNavigate();
  const submit = useSubmit();
  
  // Modal states
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [projectToDelete, setProjectToDelete] = useState<Project | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Handle page change
  const handlePageChange = (page: number) => {
    const searchParams = new URLSearchParams();
    searchParams.set('page', page.toString());
    searchParams.set('pageSize', filter.pageSize.toString());
    if (filter.search) searchParams.set('search', filter.search);
    if (filter.sortBy) searchParams.set('sortBy', filter.sortBy);
    if (filter.sortOrder) searchParams.set('sortOrder', filter.sortOrder);

    navigate(`/projects?${searchParams.toString()}`);
  };

  // Handle search
  const handleSearch = (query: string) => {
    const searchParams = new URLSearchParams();
    searchParams.set('page', '1'); // Reset to first page on search
    searchParams.set('pageSize', filter.pageSize.toString());
    if (query) searchParams.set('search', query);
    if (filter.sortBy) searchParams.set('sortBy', filter.sortBy);
    if (filter.sortOrder) searchParams.set('sortOrder', filter.sortOrder);

    navigate(`/projects?${searchParams.toString()}`);
  };

  // Handle delete - open modal instead of browser confirm
  const handleDelete = (id: string) => {
    const project = projectsData.projects.find(p => p.id === id);
    if (project) {
      setProjectToDelete(project);
      setIsModalOpen(true);
      setDeleteError(null);
    }
  };

  // Handle modal close
  const handleModalClose = () => {
    if (!isDeleting) {
      setIsModalOpen(false);
      setProjectToDelete(null);
      setDeleteError(null);
    }
  };

  // Handle actual delete confirmation
  const handleDeleteConfirm = async () => {
    if (!projectToDelete) return;
    
    setIsDeleting(true);
    setDeleteError(null);
    
    try {
      // Use Form submission with hidden input to call server action
      const formData = new FormData();
      formData.set('projectId', projectToDelete.id);
      formData.set('_action', 'delete');
      
      // Submit using the form action that will be handled by the action function
      submit(formData, { method: 'post' });
      
      // Explicitly close modal and reset state after form submission
      setIsModalOpen(false);
      setProjectToDelete(null);
      setIsDeleting(false);
    } catch (error) {
      console.error('Error deleting project:', error);
      setDeleteError('Failed to delete project. Please try again.');
      setIsDeleting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration - using bolder dot pattern */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay with more vibrant colors */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-6">
          {/* Header Section */}
          <div className="text-center space-y-3">
            <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
              Your Projects
            </h1>
            <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
              Manage and deploy your applications with enterprise-grade reliability
            </p>
          </div>

          {/* Status Messages */}
          {error && (
            <div className="max-w-4xl mx-auto">
              <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <svg className="h-4 w-4 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Error</h3>
                    <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            </div>
          )}
          
          {deleteError && (
            <div className="max-w-4xl mx-auto">
              <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <svg className="h-4 w-4 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Deletion Error</h3>
                    <p className="text-sm text-red-700 dark:text-red-300 mt-1">{deleteError}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Main Content */}
          <div className="relative">
            <ProjectList
              projects={projectsData.projects}
              totalCount={projectsData.totalCount}
              currentPage={projectsData.page}
              pageSize={projectsData.pageSize}
              totalPages={projectsData.totalPages}
              onPageChange={handlePageChange}
              onSearch={handleSearch}
              onDelete={handleDelete}
            />
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <DeleteConfirmationModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        onConfirm={handleDeleteConfirm}
        isDeleting={isDeleting}
        projectName={projectToDelete?.name || ''}
      />
    </div>
  );
}