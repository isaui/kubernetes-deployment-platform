import { useState, useEffect, useRef } from "react";
import { Environment } from "~/types/project";
import { createEnvironment, updateEnvironment } from "~/actions/environment";
import { X, Save, Database, AlertCircle, Loader2 } from "lucide-react";

interface EnvironmentModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  environment?: Environment;
  onSuccess: () => void;
}

export default function EnvironmentModal({
  isOpen,
  onClose,
  projectId,
  environment,
  onSuccess,
}: EnvironmentModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);
  const nameInputRef = useRef<HTMLInputElement>(null);

  // Set initial values if editing an existing environment
  useEffect(() => {
    if (environment) {
      setName(environment.name);
      setDescription(environment.description || "");
    } else {
      setName("");
      setDescription("");
    }
    setError(null);
  }, [environment, isOpen]);

  // Handle ESC key and focus management
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isSubmitting) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // Focus on name input when modal opens
      setTimeout(() => {
        nameInputRef.current?.focus();
      }, 100);
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen, isSubmitting, onClose]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }

    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError("Environment name is required");
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      if (environment) {
        // Update existing environment
        await updateEnvironment(environment.id, { name, description, projectId });
      } else {
        // Create new environment
        await createEnvironment({ name, description, projectId });
      }
      
      onSuccess();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "An unknown error occurred");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black/40 backdrop-blur-sm transition-opacity"
        onClick={!isSubmitting ? onClose : undefined}
      />
      
      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div
          ref={modalRef}
          className="relative bg-white/95 dark:bg-gray-800/95 backdrop-blur-md rounded-2xl shadow-xl border border-gray-200/50 dark:border-gray-700/50 w-full max-w-md transform transition-all"
          style={{
            animation: isOpen ? 'modalSlideIn 0.2s ease-out' : undefined
          }}
        >
          {/* Close button */}
          {!isSubmitting && (
            <button
              onClick={onClose}
              className="absolute right-4 top-4 p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            >
              <X size={18} />
            </button>
          )}

          {/* Content */}
          <div className="p-6">
            {/* Header */}
            <div className="text-center space-y-4 mb-6">
              {/* Icon */}
              <div className="mx-auto w-16 h-16 bg-gradient-to-br from-indigo-100 to-indigo-200 dark:from-indigo-900/50 dark:to-indigo-800/50 rounded-2xl flex items-center justify-center">
                <Database className="w-8 h-8 text-indigo-600 dark:text-indigo-400" />
              </div>
              
              {/* Title */}
              <div>
                <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                  {environment ? "Edit Environment" : "Create Environment"}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {environment 
                    ? "Update your environment configuration" 
                    : "Set up a new environment for your services"
                  }
                </p>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Error Message */}
              {error && (
                <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
                  <div className="flex items-start">
                    <div className="flex-shrink-0">
                      <AlertCircle className="h-4 w-4 text-red-500 mt-0.5" />
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Validation Error</h3>
                      <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
                    </div>
                  </div>
                </div>
              )}
              
              {/* Name Field */}
              <div className="space-y-2">
                <label htmlFor="name" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Environment Name <span className="text-red-500">*</span>
                </label>
                <input
                  ref={nameInputRef}
                  type="text"
                  id="name"
                  name="name"
                  required
                  disabled={isSubmitting}
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value);
                    if (error) setError(null); // Clear error on input
                  }}
                  className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                  placeholder="e.g. production, staging, development"
                />
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Choose a descriptive name for this environment
                </p>
              </div>
              
              {/* Description Field */}
              <div className="space-y-2">
                <label htmlFor="description" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Description
                  <span className="ml-1 text-xs font-normal text-gray-500 dark:text-gray-400">(Optional)</span>
                </label>
                <textarea
                  id="description"
                  name="description"
                  rows={3}
                  disabled={isSubmitting}
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm resize-none disabled:opacity-50"
                  placeholder="Describe the purpose of this environment..."
                />
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Add details about when and how this environment should be used
                </p>
              </div>
              
              {/* Action Buttons */}
              <div className="flex flex-col-reverse sm:flex-row gap-3 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
                <button
                  type="button"
                  onClick={onClose}
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
                      {environment ? 'Updating...' : 'Creating...'}
                    </>
                  ) : (
                    <>
                      <Save size={16} className="mr-2" />
                      {environment ? 'Update Environment' : 'Create Environment'}
                    </>
                  )}
                </button>
              </div>
            </form>

            {/* Additional Info */}
            <div className="text-center pt-4 border-t border-gray-200/50 dark:border-gray-700/50 mt-6">
              <p className="text-xs text-gray-500 dark:text-gray-400">
                {environment 
                  ? 'Changes will be applied immediately after updating' 
                  : 'You can deploy services to this environment after creation'
                }
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Custom animation styles */}
      <style>{`
        @keyframes modalSlideIn {
          from {
            opacity: 0;
            transform: scale(0.95) translateY(-10px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }
      `}</style>
    </div>
  );
}