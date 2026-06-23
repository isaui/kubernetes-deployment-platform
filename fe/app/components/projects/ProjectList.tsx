import { useState } from 'react';
import { Search, Plus, Grid, List, Folder, Sparkles } from 'lucide-react';
import type { Project } from '~/types/project';
import ProjectCard from './ProjectCard';
import ProjectTable from '~/components/projects/ProjectTable';
import { Link } from '@remix-run/react';

interface ProjectListProps {
  projects: Project[];
  totalCount: number;
  currentPage: number;
  pageSize: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  onSearch: (query: string) => void;
  onDelete: (id: string) => void;
}

export default function ProjectList({
  projects = [],
  totalCount,
  currentPage,
  pageSize,
  totalPages,
  onPageChange,
  onSearch,
  onDelete
}: ProjectListProps) {
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(searchQuery);
  };

  return (
    <div className="space-y-8">
      {/* Controls Header */}
      <div className="bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 p-6 shadow-sm">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0">
          {/* Search */}
          <div className="flex-1 max-w-lg">
            <form onSubmit={handleSearch} className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <Search className="h-5 w-5 text-gray-400 group-focus-within:text-indigo-500 transition-colors" />
              </div>
              <input
                type="text"
                placeholder="Search projects..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-12 pr-4 py-3 bg-gray-50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all"
              />
              <button 
                type="submit"
                className="absolute inset-y-0 right-0 pr-4 flex items-center text-gray-400 hover:text-indigo-600 transition-colors"
              >
                <kbd className="hidden sm:inline-flex items-center px-2 py-1 text-xs font-mono bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 rounded border border-gray-200 dark:border-gray-600">⏎</kbd>
              </button>
            </form>
          </div>

          {/* View Toggle & Create Button */}
          <div className="flex items-center space-x-4">
            {/* View Mode Toggle */}
            <div className="flex items-center bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-md transition-all ${
                  viewMode === 'grid'
                    ? 'bg-white dark:bg-gray-600 text-indigo-600 dark:text-indigo-400 shadow-sm'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
                }`}
              >
                <Grid size={18} />
              </button>
              <button
                onClick={() => setViewMode('table')}
                className={`p-2 rounded-md transition-all ${
                  viewMode === 'table'
                    ? 'bg-white dark:bg-gray-600 text-indigo-600 dark:text-indigo-400 shadow-sm'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
                }`}
              >
                <List size={18} />
              </button>
            </div>

            {/* Create Project Button */}
            <Link 
              to="/projects/new"
              className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105"
            >
              <Plus size={18} className="mr-2" />
              <span>New Project</span>
            </Link>
          </div>
        </div>

        {/* Stats */}
        {totalCount > 0 && (
          <div className="mt-4 flex items-center text-sm text-gray-600 dark:text-gray-400">
            <Folder size={16} className="mr-2" />
            <span>
              {totalCount} project{totalCount !== 1 ? 's' : ''} 
              {searchQuery && ` matching "${searchQuery}"`}
            </span>
          </div>
        )}
      </div>

      {/* Content */}
      {projects.length > 0 ? (
        <div className="space-y-8">
          {/* Projects Grid/Table */}
          {viewMode === 'grid' ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {projects.map((project) => (
                <ProjectCard key={project.id} project={project} onDelete={onDelete} />
              ))}
            </div>
          ) : (
            <div className="bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 overflow-hidden shadow-sm">
              <ProjectTable projects={projects} onDelete={onDelete} />
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center">
              <nav className="inline-flex items-center bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-2">
                {/* Previous Button */}
                <button
                  onClick={() => onPageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className={`relative inline-flex items-center px-4 py-2 text-sm font-medium rounded-xl transition-all ${
                    currentPage === 1
                      ? 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  <svg className="h-4 w-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                  Previous
                </button>
                
                {/* Page Numbers */}
                <div className="flex items-center px-2">
                  {Array.from({ length: totalPages }).map((_, index) => {
                    const pageNumber = index + 1;
                    const isCurrentPage = pageNumber === currentPage;
                    
                    // Only show a window of 5 pages
                    if (
                      pageNumber === 1 ||
                      pageNumber === totalPages ||
                      (pageNumber >= currentPage - 1 && pageNumber <= currentPage + 1)
                    ) {
                      return (
                        <button
                          key={pageNumber}
                          onClick={() => onPageChange(pageNumber)}
                          className={`relative inline-flex items-center px-4 py-2 text-sm font-medium rounded-xl transition-all mx-1 ${
                            isCurrentPage
                              ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/25'
                              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                          }`}
                        >
                          {pageNumber}
                        </button>
                      );
                    }
                    
                    // Add ellipsis indicators
                    if (
                      (pageNumber === 2 && currentPage > 3) ||
                      (pageNumber === totalPages - 1 && currentPage < totalPages - 2)
                    ) {
                      return (
                        <span
                          key={`ellipsis-${pageNumber}`}
                          className="relative inline-flex items-center px-4 py-2 text-sm font-medium text-gray-500 dark:text-gray-400"
                        >
                          ...
                        </span>
                      );
                    }
                    
                    return null;
                  })}
                </div>
                
                {/* Next Button */}
                <button
                  onClick={() => onPageChange(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className={`relative inline-flex items-center px-4 py-2 text-sm font-medium rounded-xl transition-all ${
                    currentPage === totalPages
                      ? 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  Next
                  <svg className="h-4 w-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </nav>
            </div>
          )}
          
          {/* Results Summary */}
          <div className="text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Showing <span className="font-semibold text-gray-900 dark:text-white">{projects.length}</span> of{' '}
              <span className="font-semibold text-gray-900 dark:text-white">{totalCount}</span> projects
            </p>
          </div>
        </div>
      ) : (
        /* Empty State */
        <div className="relative">
          <div className="bg-gradient-to-br from-white/80 to-indigo-50/50 dark:from-gray-800/80 dark:to-indigo-900/20 backdrop-blur-sm rounded-3xl border border-gray-200/50 dark:border-gray-700/50 p-12 text-center shadow-xl">
            {/* Background decoration */}
            <div className="absolute inset-0 bg-gradient-to-br from-indigo-50/50 to-purple-50/30 dark:from-indigo-900/20 dark:to-purple-900/10 rounded-3xl" />
            
            <div className="relative">
              <div className="mx-auto w-24 h-24 bg-gradient-to-br from-indigo-100 to-indigo-200 dark:from-indigo-900/50 dark:to-indigo-800/50 rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                {searchQuery ? (
                  <Search className="h-10 w-10 text-indigo-600 dark:text-indigo-400" />
                ) : (
                  <div className="relative">
                    <Folder className="h-10 w-10 text-indigo-600 dark:text-indigo-400" />
                    <Sparkles className="h-4 w-4 text-indigo-500 absolute -top-1 -right-1 animate-pulse" />
                  </div>
                )}
              </div>
              
              <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                {searchQuery ? 'No projects found' : 'Ready to deploy?'}
              </h3>
              
              <p className="text-lg text-gray-600 dark:text-gray-300 mb-8 max-w-md mx-auto">
                {searchQuery 
                  ? `No projects match "${searchQuery}". Try adjusting your search terms.`
                  : 'Create your first project and start deploying applications with Kubesa.'
                }
              </p>
              
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Link
                  to="/projects/new"
                  className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105"
                >
                  <Plus size={20} className="mr-3" />
                  Create Your First Project
                </Link>
                
                {searchQuery && (
                  <button
                    onClick={() => {
                      setSearchQuery('');
                      onSearch('');
                    }}
                    className="inline-flex items-center px-8 py-4 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-semibold rounded-xl border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200"
                  >
                    Clear Search
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}