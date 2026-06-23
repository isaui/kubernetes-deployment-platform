import { useState } from 'react';
import { X, Plus, Save, AlertCircle } from 'lucide-react';

interface EnvironmentVariableFormProps {
  initialName?: string;
  initialValue?: string;
  isEditing?: boolean;
  onSubmit: (name: string, value: string) => void;
  onCancel: () => void;
}

export default function EnvironmentVariableForm({
  initialName = '',
  initialValue = '',
  isEditing = false,
  onSubmit,
  onCancel,
}: EnvironmentVariableFormProps) {
  const [name, setName] = useState(initialName);
  const [value, setValue] = useState(initialValue);
  const [nameError, setNameError] = useState('');
  const [valueError, setValueError] = useState('');

  const validateForm = () => {
    let isValid = true;
    
    // Reset errors
    setNameError('');
    setValueError('');
    
    // Validate name
    if (!name.trim()) {
      setNameError('Name is required');
      isValid = false;
    } else if (!/^[A-Z0-9_]+$/i.test(name)) {
      setNameError('Name can only contain letters, numbers, and underscores');
      isValid = false;
    }
    
    // Validate value (optional validation)
    if (!value.trim()) {
      setValueError('Value is required');
      isValid = false;
    }
    
    return isValid;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (validateForm()) {
      onSubmit(name, value);
      // Reset form if not editing
      if (!isEditing) {
        setName('');
        setValue('');
      }
    }
  };

  return (
    <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            {isEditing ? 'Edit Variable' : 'Add New Variable'}
          </h3>
          <button
            type="button"
            onClick={onCancel}
            className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="p-6 space-y-6">
        {/* Name Field */}
        <div>
          <label htmlFor="varName" className="block text-sm font-semibold text-gray-900 dark:text-white mb-2">
            Variable Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="varName"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              if (nameError) setNameError(''); // Clear error on input
            }}
            disabled={isEditing}
            placeholder="e.g., API_KEY, DATABASE_URL"
            className={`w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm font-mono ${
              nameError 
                ? 'border-red-300 dark:border-red-600 focus:border-red-500 dark:focus:border-red-400' 
                : 'border-gray-200 dark:border-gray-700 focus:border-indigo-500 dark:focus:border-indigo-400'
            } ${isEditing ? 'opacity-60 cursor-not-allowed' : ''}`}
          />
          {nameError && (
            <div className="mt-2 flex items-center space-x-2 text-red-600 dark:text-red-400">
              <AlertCircle size={14} />
              <p className="text-xs">{nameError}</p>
            </div>
          )}
          {!nameError && (
            <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
              Use uppercase letters, numbers, and underscores only
            </p>
          )}
        </div>
        
        {/* Value Field */}
        <div>
          <label htmlFor="varValue" className="block text-sm font-semibold text-gray-900 dark:text-white mb-2">
            Variable Value <span className="text-red-500">*</span>
          </label>
          <textarea
            id="varValue"
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              if (valueError) setValueError(''); // Clear error on input
            }}
            rows={3}
            placeholder="Enter the variable value..."
            className={`w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm resize-none font-mono ${
              valueError 
                ? 'border-red-300 dark:border-red-600 focus:border-red-500 dark:focus:border-red-400' 
                : 'border-gray-200 dark:border-gray-700 focus:border-indigo-500 dark:focus:border-indigo-400'
            }`}
          />
          {valueError && (
            <div className="mt-2 flex items-center space-x-2 text-red-600 dark:text-red-400">
              <AlertCircle size={14} />
              <p className="text-xs">{valueError}</p>
            </div>
          )}
          {!valueError && (
            <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
              This value will be available to your application at runtime
            </p>
          )}
        </div>
        
        {/* Action Buttons */}
        <div className="flex flex-col-reverse sm:flex-row gap-3 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 sm:flex-none px-6 py-3 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
          >
            Cancel
          </button>
          
          <button
            type="submit"
            className="flex-1 sm:flex-none px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 transform hover:scale-105 flex items-center justify-center"
          >
            {isEditing ? (
              <>
                <Save size={16} className="mr-2" />
                Update Variable
              </>
            ) : (
              <>
                <Plus size={16} className="mr-2" />
                Add Variable
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}