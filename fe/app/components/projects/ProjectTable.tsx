import { Link } from '@remix-run/react';
import { Edit, Trash, Folder, ExternalLink } from 'lucide-react';
import { useState } from 'react';
import type { Project } from '~/types/project';

interface ProjectTableProps {
  projects: Project[];
  onDelete: (id: string) => void;
}

export default function ProjectTable({ projects, onDelete }: ProjectTableProps) {
  // Date formatting removed

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Project
            </th>
            <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Status
            </th>
            <th className="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
          {projects.map((project) => (
            <tr 
              key={project.id} 
              className="group hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors"
            >
              {/* Project Info */}
              <td className="px-6 py-4">
                <div className="flex items-center space-x-4">
                  {/* Project Icon */}
                  <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-indigo-100 to-indigo-200 dark:from-indigo-900/50 dark:to-indigo-800/50 rounded-lg flex items-center justify-center">
                    <Folder className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  
                  {/* Project Details */}
                  <div className="min-w-0 flex-1">
                    <Link 
                      to={`/projects/${project.id}`}
                      className="group/link flex items-center space-x-2"
                    >
                      <span className="text-sm font-semibold text-gray-900 dark:text-white group-hover/link:text-indigo-600 dark:group-hover/link:text-indigo-400 transition-colors">
                        {project.name}
                      </span>
                      <ExternalLink className="w-3 h-3 text-gray-400 opacity-0 group-hover/link:opacity-100 transition-opacity" />
                    </Link>
                    {project.description && (
                      <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1 max-w-xs">
                        {project.description}
                      </p>
                    )}
                  </div>
                </div>
              </td>

              {/* Status */}
              <td className="px-6 py-4">
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                  <span className="text-sm font-medium text-green-700 dark:text-green-400">Ready</span>
                </div>
              </td>

              {/* Actions */}
              <td className="px-6 py-4 text-right">
                <div className="flex items-center justify-end space-x-2">
                  <Link
                    to={`/projects/${project.id}/edit`}
                    className="p-2 text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-all"
                    title="Edit Project"
                  >
                    <Edit size={16} />
                  </Link>
                  <button
                    onClick={() => onDelete(project.id)}
                    className="p-2 text-red-500 hover:text-red-600 dark:text-red-400 dark:hover:text-red-300 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-all"
                    title="Delete Project"
                  >
                    <Trash size={16} />
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Empty State for Table */}
      {projects.length === 0 && (
        <div className="text-center py-12">
          <div className="mx-auto w-16 h-16 bg-gradient-to-br from-indigo-100 to-indigo-200 dark:from-indigo-900/50 dark:to-indigo-800/50 rounded-xl flex items-center justify-center mb-4">
            <Folder className="w-8 h-8 text-indigo-600 dark:text-indigo-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">No projects found</h3>
          <p className="text-gray-500 dark:text-gray-400">Start by creating your first project</p>
        </div>
      )}
    </div>
  );
}