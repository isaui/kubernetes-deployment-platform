import { Link } from '@remix-run/react';
import { Edit, MoreVertical, Trash, ArrowRight, Folder } from 'lucide-react';
import { useState } from 'react';
import type { Project } from '~/types/project';

interface ProjectCardProps {
  project: Project;
  onDelete: (id: string) => void;
}

export default function ProjectCard({ project, onDelete }: ProjectCardProps) {
  const [showDropdown, setShowDropdown] = useState(false);

  // Date formatting removed

  return (
    <div className="group relative">
      {/* Gradient background blur effect */}
      <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500 to-purple-600 rounded-2xl blur opacity-0 group-hover:opacity-20 transition duration-300" />
      
      <div className="relative bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 p-6 transition-all duration-300 hover:shadow-xl hover:shadow-indigo-500/10 hover:-translate-y-1">
        {/* Header */}
        <div className="flex justify-between items-start mb-4">
          <div className="flex items-start space-x-3 flex-1">
            {/* Project Icon */}
            <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-indigo-100 to-indigo-200 dark:from-indigo-900/50 dark:to-indigo-800/50 rounded-xl flex items-center justify-center shadow-sm group-hover:shadow-md transition-shadow">
              <Folder className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
            </div>
            
            {/* Project Name */}
            <div className="flex-1 min-w-0">
              <Link to={`/projects/${project.id}`} className="group/title">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white group-hover/title:text-indigo-600 dark:group-hover/title:text-indigo-400 transition-colors duration-200 line-clamp-1">
                  {project.name}
                </h3>
              </Link>
            </div>
          </div>
          
          {/* Actions Menu */}
          <div className="relative flex-shrink-0">
            <button
              onClick={() => setShowDropdown(!showDropdown)}
              className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-all opacity-0 group-hover:opacity-100"
            >
              <MoreVertical size={18} />
            </button>
            
            {showDropdown && (
              <>
                <div 
                  className="fixed inset-0 z-40" 
                  onClick={() => setShowDropdown(false)}
                />
                <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 z-50 overflow-hidden backdrop-blur-sm">
                  <div className="py-2">
                    <Link 
                      to={`/projects/${project.id}/edit`}
                      className="flex items-center px-4 py-3 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
                      onClick={() => setShowDropdown(false)}
                    >
                      <Edit size={16} className="mr-3 text-gray-500" />
                      Edit Project
                    </Link>
                    <button
                      onClick={() => {
                        onDelete(project.id);
                        setShowDropdown(false);
                      }}
                      className="flex items-center w-full text-left px-4 py-3 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                    >
                      <Trash size={16} className="mr-3" />
                      Delete Project
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
        
        {/* Description */}
        {project.description && (
          <p className="text-gray-600 dark:text-gray-300 text-sm line-clamp-2 mb-4 leading-relaxed">
            {project.description}
          </p>
        )}
        
        {/* Metadata section removed */}
        
        {/* Footer Action */}
        <div className="flex items-center justify-between pt-4 border-t border-gray-100 dark:border-gray-700">
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Ready</span>
          </div>
          
          <Link 
            to={`/projects/${project.id}`}
            className="inline-flex items-center text-sm font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors group/link"
          >
            <span>View Details</span>
            <ArrowRight size={14} className="ml-1 transition-transform group-hover/link:translate-x-1" />
          </Link>
        </div>
        
        {/* Hover Effect Gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-50/0 via-transparent to-purple-50/0 dark:from-indigo-900/0 dark:to-purple-900/0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />
      </div>
    </div>
  );
}