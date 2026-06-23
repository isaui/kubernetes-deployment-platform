import { useState, useEffect } from 'react';
import { Plus, AlertCircle, Database, Save, Info, Code } from 'lucide-react';
import { useNavigate } from '@remix-run/react';
import EnvironmentVariableItem from './EnvironmentVariableItem';
import EnvironmentVariableForm from './EnvironmentVariableForm';
import RawEditor from './RawEditor';
import type { Service } from '~/types/service';
import { updateService } from '~/actions/service-update';
import { createServiceUpdate } from '~/types/service-update';

// Define interfaces
interface EnvironmentVariable {
  name: string;
  value: string;
}

interface EnvironmentVariablesTabProps {
  service: Service;
}

export default function EnvironmentVariablesTab({ service }: EnvironmentVariablesTabProps) {
  // Initialize Remix navigation
  const navigate = useNavigate();
  
  // Local state for environment variables
  const [variables, setVariables] = useState<EnvironmentVariable[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [showRawEditor, setShowRawEditor] = useState(false);
  const [editingVariable, setEditingVariable] = useState<EnvironmentVariable | null>(null);
  
  // Track if we've made changes that need to be saved
  const [isDirty, setIsDirty] = useState(false);
  // Store bulk operations to be submitted together
  const [pendingChanges, setPendingChanges] = useState<Record<string, string>>({});
  
  // Determine if the service type allows editing
  const isEditable = service.type === 'git';
  
  // Initialize variables from service data
  useEffect(() => {
    try {
      // Initialize from service.envVars for immediate display
      const envVarsObj = service.envVars || {};
      const envVarsArray = Object.entries(envVarsObj).map(
        ([name, value]) => ({ name, value: String(value) })
      );
      setVariables(envVarsArray);
      setPendingChanges(envVarsObj);
      setIsLoading(false);
    } catch (err) {
      console.error('Error parsing env vars from service:', err);
      setError('Error loading environment variables');
      setIsLoading(false);
    }
  }, [service.envVars]);
  
  // Submit all pending changes at once
  const submitChanges = async () => {
    if (!isEditable || !isDirty) return;
    
    setIsSaving(true);
    try {
      // Create service update with only envVars
      const serviceUpdate = createServiceUpdate(
        { ...service, envVars: pendingChanges },
        ['envVars']
      );
      
      // Call the client-side action to update environment variables
      await updateService(service.id, serviceUpdate);
      
      // Use Remix navigation to refresh the current route with new data
      // This will trigger a reload of the data without a full page refresh
      navigate(".", { replace: true });
      
      // Reset local state
      setError(null);
      setIsDirty(false);
      setShowForm(false);
      setShowRawEditor(false);
      
      // We don't need to manually update local state anymore as the component
      // will be rehydrated with fresh data from the server
    } catch (err) {
      console.error('Error saving environment variables:', err);
      setError(err instanceof Error ? err.message : 'Failed to save environment variables');
    } finally {
      setIsSaving(false);
    }
  };

  // Handle add variable
  const handleAddVariable = (name: string, value: string) => {
    if (!isEditable) return;
    
    // Update local state
    setVariables(prev => [...prev, { name, value }]);
    
    // Track changes for bulk update
    setPendingChanges(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Mark as dirty to save later
    setIsDirty(true);
    setShowForm(false);
  };
  
  // Handle edit variable
  const handleEditVariable = (name: string, value: string) => {
    if (!isEditable) return;
    
    // Update local state
    setVariables(prev => 
      prev.map(v => v.name === name ? { name, value } : v)
    );
    
    // Track changes for bulk update
    setPendingChanges(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Mark as dirty to save later
    setIsDirty(true);
    setEditingVariable(null);
  };
  
  // Handle delete variable
  const handleDeleteVariable = (name: string) => {
    if (!isEditable) return;
    
    // Update local state
    setVariables(prev => prev.filter(v => v.name !== name));
    
    // Track deletion in pending changes
    setPendingChanges(prev => {
      const updated = { ...prev };
      delete updated[name];
      return updated;
    });
    
    // Mark as dirty to save later
    setIsDirty(true);
  };
  
  // Handle raw editor submit
  const handleRawEditorSubmit = (newVariables: Record<string, string>) => {
    if (!isEditable) return;
    
    // Update local state
    const newVariablesArray = Object.entries(newVariables).map(
      ([name, value]) => ({ name, value })
    );
    setVariables(newVariablesArray);
    
    // Track changes for bulk update
    setPendingChanges(newVariables);
    
    // Mark as dirty to save later
    setIsDirty(true);
    setShowRawEditor(false);
  };
  
  // Start editing a variable
  const startEditingVariable = (name: string, value: string) => {
    setEditingVariable({ name, value });
  };
  
  // Manual save only - no auto-save
  const handleSaveNow = () => {
    if (isDirty) {
      submitChanges();
    }
  };

  // Toggle between form and raw editor
  const toggleRawEditor = () => {
    setShowRawEditor(!showRawEditor);
    setShowForm(false);
    setEditingVariable(null);
  };

  const toggleForm = () => {
    setShowForm(!showForm);
    setShowRawEditor(false);
    setEditingVariable(null);
  };

  return (
    <div className="p-8 space-y-8">
      {/* Info box for read-only variables */}
      {!isEditable && (
        <div className="bg-blue-50/80 dark:bg-blue-900/20 border border-blue-200/60 dark:border-blue-800/30 rounded-xl p-6 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <AlertCircle className="w-6 h-6 text-blue-500 dark:text-blue-400 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-2">Read-only Variables</h3>
              <p className="text-sm text-blue-700 dark:text-blue-400">
                Environment variables for managed services are read-only and automatically configured by Kubesa.
              </p>
            </div>
          </div>
        </div>
      )}
      
      {/* Error display */}
      {error && (
        <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-4 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <AlertCircle className="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" />
            <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
          </div>
        </div>
      )}
      
      {/* Loading and saving indicators */}
      {(isLoading || isSaving) && (
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-xl border border-gray-200/50 dark:border-gray-700/50 p-8 text-center">
          <div className="flex items-center justify-center space-x-3">
            <div className="animate-spin rounded-full h-6 w-6 border-2 border-indigo-600 border-t-transparent"></div>
            <p className="text-gray-600 dark:text-gray-400">
              {isSaving ? 'Saving changes...' : 'Loading variables...'}
            </p>
          </div>
        </div>
      )}
      
      {/* Unsaved changes indicator */}
      {isDirty && (
        <div className="bg-yellow-50/80 dark:bg-yellow-900/20 border border-yellow-200/60 dark:border-yellow-700/50 rounded-xl p-4 backdrop-blur-sm">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Info className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
              <p className="text-yellow-800 dark:text-yellow-200 font-medium">
                You have unsaved changes
              </p>
            </div>
            <button
              onClick={handleSaveNow}
              className="inline-flex items-center px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white font-medium rounded-lg transition-colors"
            >
              <Save className="w-4 h-4 mr-2" />
              Save Now
            </button>
          </div>
        </div>
      )}

      {/* Raw Editor */}
      {showRawEditor && (
        <RawEditor
          initialVariables={pendingChanges}
          onSubmit={handleRawEditorSubmit}
          onCancel={() => setShowRawEditor(false)}
        />
      )}

      {/* Main Content */}
      <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
        {/* Header with Add Button */}
        <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Variables</h3>
              <span className="inline-flex items-center px-2 py-1 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-xs font-medium text-indigo-700 dark:text-indigo-300">
                {variables.length}
              </span>
            </div>
            
            {isEditable && !showForm && !editingVariable && !showRawEditor && (
              <div className="flex items-center space-x-2">
                {/* Raw Editor Toggle */}
                <button
                  onClick={toggleRawEditor}
                  className="inline-flex items-center px-4 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-xl transition-all duration-200"
                  title="Raw Editor - Bulk edit with .env or JSON format"
                >
                  <Code size={16} className="mr-2" />
                  Raw Editor
                </button>
                
                {/* Add Variable Button */}
                <button
                  onClick={toggleForm}
                  className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105"
                >
                  <Plus size={16} className="mr-2" />
                  Add Variable
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Add form */}
          {showForm && (
            <div className="mb-6">
              <EnvironmentVariableForm
                onSubmit={handleAddVariable}
                onCancel={() => setShowForm(false)}
              />
            </div>
          )}
          
          {/* Edit form */}
          {editingVariable && (
            <div className="mb-6">
              <EnvironmentVariableForm
                onSubmit={(name, value) => handleEditVariable(name, value)}
                onCancel={() => setEditingVariable(null)}
                initialName={editingVariable.name}
                initialValue={editingVariable.value}
                isEditing
              />
            </div>
          )}
          
          {/* Variables list or empty state */}
          {variables.length === 0 && !isLoading ? (
            <div className="text-center py-12">
              <div className="mx-auto w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-2xl flex items-center justify-center mb-4">
                <Database className="w-8 h-8 text-gray-400" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">No Variables Defined</h3>
              <p className="text-gray-500 dark:text-gray-400 mb-4">
                {isEditable ? 'Get started by adding your first environment variable.' : 'This service has no environment variables configured.'}
              </p>
              {isEditable && (
                <div className="flex flex-col sm:flex-row gap-3 items-center justify-center">
                  <button
                    onClick={toggleForm}
                    className="inline-flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg transition-colors"
                  >
                    <Plus size={16} className="mr-2" />
                    Add Variable
                  </button>
                  <span className="text-gray-400">or</span>
                  <button
                    onClick={toggleRawEditor}
                    className="inline-flex items-center px-4 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors"
                  >
                    <Code size={16} className="mr-2" />
                    Use Raw Editor
                  </button>
                </div>
              )}
            </div>
          ) : (
            <div className="space-y-1">
              {/* Header */}
              <div className="grid grid-cols-12 gap-4 px-4 py-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                <div className="col-span-4">Name</div>
                <div className="col-span-6">Value</div>
                {isEditable && <div className="col-span-2 text-right">Actions</div>}
              </div>
              
              {/* Variables */}
              <div className="space-y-1">
                {variables.map((variable) => (
                  <EnvironmentVariableItem
                    key={variable.name}
                    name={variable.name}
                    value={variable.value}
                    isEditable={isEditable}
                    onEdit={startEditingVariable}
                    onDelete={handleDeleteVariable}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}