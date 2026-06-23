import { useState, useEffect } from 'react';
import { Form, useNavigation } from '@remix-run/react';
import { Save, X, AlertCircle, Loader2 } from 'lucide-react';
import type { Project } from '~/types/project';

interface ProjectFormProps {
  project?: Project;
  isEditing?: boolean;
}

export default function ProjectForm({ project, isEditing = false }: ProjectFormProps) {
  const navigation = useNavigation();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [formError, setFormError] = useState<string | null>(null);

  const isSubmitting = navigation.state === 'submitting';

  // Populate form when editing
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description || '');
    }
  }, [project]);

  // Form validation
  const validateForm = () => {
    if (!name.trim()) {
      setFormError('Project name is required');
      return false;
    }
    if (name.trim().length < 2) {
      setFormError('Project name must be at least 2 characters long');
      return false;
    }
    if (name.trim().length > 50) {
      setFormError('Project name must be less than 50 characters');
      return false;
    }
    setFormError(null);
    return true;
  };

  const handleSubmit = (e: React.FormEvent) => {
    if (!validateForm()) {
      e.preventDefault();
    }
  };

  return (
    <div className="space-y-8">
      {/* Form Header */}
      <div className="text-center space-y-2">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          {isEditing ? 'Update Project Details' : 'Project Information'}
        </h2>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {isEditing 
            ? 'Modify the project name and description below' 
            : 'Fill in the basic information for your new project'
          }
        </p>
      </div>

      <Form method="post" onSubmit={handleSubmit} className="space-y-6">
        {/* Form Error */}
        {formError && (
          <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <AlertCircle className="h-4 w-4 text-red-500 mt-0.5" />
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Validation Error</h3>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">{formError}</p>
              </div>
            </div>
          </div>
        )}

        {/* Project Name */}
        <div className="space-y-2">
          <label htmlFor="name" className="block text-sm font-semibold text-gray-900 dark:text-white">
            Project Name <span className="text-red-500">*</span>
          </label>
          <div className="relative">
            <input
              type="text"
              id="name"
              name="name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (formError) setFormError(null); // Clear error on input
              }}
              className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm"
              placeholder="Enter a descriptive project name"
              required
              disabled={isSubmitting}
            />
            <div className="absolute top-3 right-3 text-xs text-gray-400">
              {name.length}/50
            </div>
          </div>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            Choose a unique name that describes your project's purpose
          </p>
        </div>

        {/* Project Description */}
        <div className="space-y-2">
          <label htmlFor="description" className="block text-sm font-semibold text-gray-900 dark:text-white">
            Description
            <span className="ml-1 text-xs font-normal text-gray-500 dark:text-gray-400">(Optional)</span>
          </label>
          <textarea
            id="description"
            name="description"
            rows={4}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm resize-none"
            placeholder="Describe what this project is for and what it will contain..."
            disabled={isSubmitting}
          />
          <p className="text-xs text-gray-500 dark:text-gray-400">
            Provide additional context about this project for your team
          </p>
        </div>

        {/* Hidden field for project ID when editing */}
        {isEditing && project && <input type="hidden" name="id" value={project.id} />}

        {/* Action Buttons */}
        <div className="flex flex-col-reverse sm:flex-row gap-4 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
          <button
            type="button"
            onClick={() => window.history.back()}
            disabled={isSubmitting}
            className="flex-1 sm:flex-none px-6 py-3 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
          >
            <X size={16} className="mr-2" />
            Cancel
          </button>
          
          <button
            type="submit"
            disabled={isSubmitting || !name.trim()}
            className="flex-1 sm:flex-none px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center transform hover:scale-105 disabled:hover:scale-100"
          >
            {isSubmitting ? (
              <>
                <Loader2 size={16} className="mr-2 animate-spin" />
                {isEditing ? 'Updating...' : 'Creating...'}
              </>
            ) : (
              <>
                <Save size={16} className="mr-2" />
                {isEditing ? 'Update Project' : 'Create Project'}
              </>
            )}
          </button>
        </div>
      </Form>

      {/* Additional Info */}
      <div className="text-center pt-4 border-t border-gray-200/50 dark:border-gray-700/50">
        <p className="text-xs text-gray-500 dark:text-gray-400">
          {isEditing 
            ? 'Changes will be saved immediately after updating' 
            : 'You can add services and configure environments after creating the project'
          }
        </p>
      </div>
    </div>
  );
}