import { json, redirect, type ActionFunctionArgs } from '@remix-run/node';
import { useActionData, Link } from '@remix-run/react';
import { createProject } from '~/actions/project.server';
import ProjectForm from '~/components/projects/ProjectForm';
import { ArrowLeft, Plus, Sparkles } from 'lucide-react';

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const name = formData.get('name') as string;
  const description = formData.get('description') as string;

  if (!name?.trim()) {
    return json(
      { error: 'Project name is required' },
      { status: 400 }
    );
  }

  try {
    const newProject = await createProject({
      name,
      description: description || undefined
    }, request);

    // Redirect to the new project detail page
    return redirect(`/projects/${newProject.id}`);
  } catch (error) {
    console.error('Error creating project:', error);
    return json(
      { error: 'Failed to create project. Please try again.' },
      { status: 500 }
    );
  }
}

export default function NewProject() {
  const actionData = useActionData<typeof action>();
  const error = actionData?.error;

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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

              
              <div className="relative">
                <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
                  Create New Project
                </h1>
                <Sparkles className="absolute -top-2 -right-8 w-6 h-6 text-indigo-500 animate-pulse hidden sm:block" />
              </div>
              
              <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                Start a new deployment project to host and manage your applications
              </p>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="max-w-2xl mx-auto">
              <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <svg className="h-4 w-4 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Creation Failed</h3>
                    <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Form Container */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl p-8">
              <ProjectForm />
            </div>
          </div>

          {/* Tips Section */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-indigo-50/60 dark:bg-indigo-900/20 backdrop-blur-sm rounded-xl border border-indigo-200/50 dark:border-indigo-700/50 p-6">
              <h3 className="text-sm font-semibold text-indigo-800 dark:text-indigo-200 mb-3">💡 Pro Tips</h3>
              <ul className="text-sm text-indigo-700 dark:text-indigo-300 space-y-2">
                <li>• Choose a descriptive name that reflects your project's purpose</li>
                <li>• Add a detailed description to help team members understand the project</li>
                <li>• You can always edit these details later from the project settings</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}