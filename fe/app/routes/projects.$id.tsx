import { json, type LoaderFunctionArgs } from '@remix-run/node';
import { useLoaderData, Link } from '@remix-run/react';
import { useState, useEffect } from 'react';
import { getProjectStats } from '~/actions/project.server';
import ProjectStats from '~/components/projects/ProjectStats';
import EnvironmentModal from '~/components/environments/EnvironmentModal';
import { Environment, ProjectEnvironmentItem } from '~/types/project';
import { ArrowLeft, Plus, Loader2, BarChart3 } from 'lucide-react';

export async function loader({ params, request }: LoaderFunctionArgs) {
  const { id } = params;
  
  if (!id) {
    throw new Response('Project ID is required', { status: 400 });
  }

  try {
    const stats = await getProjectStats(id, request);
    return json({ stats });
  } catch (error) {
    console.error('Error loading project stats:', error);
    throw new Response('Failed to load project details', { status: 500 });
  }
}

export default function ProjectDetail() {
  const { stats } = useLoaderData<typeof loader>();
  const [loading, setLoading] = useState(false);
  const [isEnvironmentModalOpen, setIsEnvironmentModalOpen] = useState(false);
  const [currentEnvironment, setCurrentEnvironment] = useState<Environment | undefined>();

  // Auto refresh stats for services with in-progress deployments
  useEffect(() => {
    let intervalId: NodeJS.Timeout | null = null;
    
    if (stats.deployments.inProgress > 0) {
      intervalId = setInterval(() => {
        setLoading(true);
        // This will trigger a refresh of the page data
        window.location.reload();
      }, 10000); // Refresh every 10 seconds
    }
    
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [stats.deployments.inProgress]);

  // Function to open create environment modal
  const openCreateEnvironmentModal = () => {
    setCurrentEnvironment(undefined);
    setIsEnvironmentModalOpen(true);
  };

  // Function to open edit environment modal
  const openEditEnvironmentModal = (environment: ProjectEnvironmentItem) => {
    // Convert ProjectEnvironmentItem to Environment by adding missing required fields
    const fullEnvironment: Environment = {
      id: environment.id,
      name: environment.name,
      description: environment.description || '',
      projectId: stats.project.id,
      createdAt: environment.createdAt,
      updatedAt: new Date().toISOString(), // Default to current time
    };
    setCurrentEnvironment(fullEnvironment);
    setIsEnvironmentModalOpen(true);
  };
  
  // Function to handle successful environment creation/update
  const handleEnvironmentSuccess = () => {
    // Refresh the page to show updated environments
    window.location.reload();
  };

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
              to="/projects"
              className="inline-flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors group"
            >
              <ArrowLeft size={16} className="mr-2 transition-transform group-hover:-translate-x-1" />
              Back to Projects
            </Link>
            
            <div className="text-center space-y-3">
              <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
                {stats.project.name}
              </h1>
              {stats.project.description && (
                <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                  {stats.project.description}
                </p>
              )}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-center">
            <div className="bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 p-4 shadow-sm">
              <div className="flex flex-wrap gap-3 justify-center">
                <Link
                  to={`/projects/${stats.project.id}/edit`}
                  className="inline-flex items-center px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                >
                  Edit Project
                </Link>
                
                <Link
                  to={`/projects/${stats.project.id}/services/create`}
                  className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl transition-all duration-200 transform hover:scale-105"
                >
                  Add Service
                </Link>
                
                <button
                  onClick={openCreateEnvironmentModal}
                  className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white font-semibold rounded-xl shadow-lg shadow-green-500/25 hover:shadow-green-500/40 transition-all duration-200 transform hover:scale-105"
                >
                  <Plus size={16} className="mr-2" />
                  Add Environment
                </button>
              </div>
            </div>
          </div>

          {/* Loading State */}
          {loading ? (
            <div className="flex justify-center items-center h-64">
              <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 p-8">
                <div className="flex items-center space-x-3">
                  <Loader2 className="h-8 w-8 text-indigo-600 animate-spin" />
                  <span className="text-lg font-medium text-gray-900 dark:text-white">
                    Refreshing data...
                  </span>
                </div>
              </div>
            </div>
          ) : (
            <ProjectStats 
              stats={stats} 
              onEditEnvironment={openEditEnvironmentModal}
              onDeleteSuccess={handleEnvironmentSuccess}
            />
          )}
        </div>
      </div>
      
      {/* Environment Modal */}
      <EnvironmentModal
        isOpen={isEnvironmentModalOpen}
        onClose={() => setIsEnvironmentModalOpen(false)}
        projectId={stats.project.id}
        environment={currentEnvironment}
        onSuccess={handleEnvironmentSuccess}
      />
    </div>
  );
}
