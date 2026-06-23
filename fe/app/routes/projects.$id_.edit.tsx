import { json, redirect, type ActionFunctionArgs, type LoaderFunctionArgs } from '@remix-run/node';
import { useLoaderData, useActionData, Link } from '@remix-run/react';
import { getProject, updateProject } from '~/actions/project.server';
import ProjectForm from '~/components/projects/ProjectForm';
import { ArrowLeft, Edit3 } from 'lucide-react';

export async function loader({ params, request }: LoaderFunctionArgs) {
  const { id } = params;
  
  if (!id) {
    throw new Response('Project ID is required', { status: 400 });
  }

  try {
    const project = await getProject(id, request);
    return json({ project });
  } catch (error) {
    console.error('Error loading project:', error);
    throw new Response('Failed to load project details', { status: 500 });
  }
}

export async function action({ request, params }: ActionFunctionArgs) {
  const { id } = params;
  
  if (!id) {
    return json(
      { error: 'Project ID is required' },
      { status: 400 }
    );
  }

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
    await updateProject(id, {
      name,
      description: description || undefined
    }, request);

    // Redirect to the project detail page
    return redirect(`/projects/${id}`);
  } catch (error) {
    console.error('Error updating project:', error);
    return json(
      { error: 'Failed to update project. Please try again.' },
      { status: 500 }
    );
  }
}

export default function EditProject() {
  const { project } = useLoaderData<typeof loader>();
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
              to={`/projects/${project.id}`}
              className="inline-flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors group"
            >
              <ArrowLeft size={16} className="mr-2 transition-transform group-hover:-translate-x-1" />
              Back to Project
            </Link>
            
            <div className="text-center space-y-3">

              <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
                Update "{project.name}"
              </h1>
              <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                Modify your project details and configuration
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
                    <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Update Failed</h3>
                    <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Form Container */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl p-8">
              <ProjectForm project={project} isEditing={true} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
