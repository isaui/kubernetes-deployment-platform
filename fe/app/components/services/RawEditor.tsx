import { useState, useEffect } from 'react';
import { Code, Save, X, AlertCircle, Info, Upload } from 'lucide-react';
import { 
  parseEnvContent, 
  parseJsonContent, 
  variablesToEnvFormat, 
  variablesToJsonFormat,
  detectFormat,
  validateEnvironmentVariables
} from '~/utils/env-parser';

interface RawEditorProps {
  initialVariables: Record<string, string>;
  onSubmit: (variables: Record<string, string>) => void;
  onCancel: () => void;
}

export default function RawEditor({ initialVariables, onSubmit, onCancel }: RawEditorProps) {
  const [content, setContent] = useState('');
  const [format, setFormat] = useState<'env' | 'json'>('env');
  const [error, setError] = useState<string | null>(null);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [previewVariables, setPreviewVariables] = useState<Record<string, string>>({});

  // Initialize content from existing variables
  useEffect(() => {
    if (format === 'env') {
      setContent(variablesToEnvFormat(initialVariables));
    } else {
      setContent(variablesToJsonFormat(initialVariables));
    }
  }, [initialVariables, format]);

  // Parse content and update preview
  useEffect(() => {
    parseContent();
  }, [content, format]);

  const parseContent = () => {
    setError(null);
    setWarnings([]);
    
    if (!content.trim()) {
      setPreviewVariables({});
      return;
    }

    let result;
    if (format === 'env') {
      result = parseEnvContent(content);
    } else {
      result = parseJsonContent(content);
    }

    if (result.errors.length > 0) {
      setError(result.errors[0]); // Show first error
      setPreviewVariables({});
    } else {
      setPreviewVariables(result.variables);
      setWarnings(result.warnings);
      
      // Additional validation
      const validation = validateEnvironmentVariables(result.variables);
      if (validation.warnings.length > 0) {
        setWarnings(prev => [...prev, ...validation.warnings]);
      }
    }
  };

  const handleFormatChange = (newFormat: 'env' | 'json') => {
    try {
      // Convert current content to new format using utility functions
      let currentVariables: Record<string, string> = {};
      
      if (format === 'env') {
        const result = parseEnvContent(content);
        if (result.errors.length === 0) {
          currentVariables = result.variables;
        }
      } else {
        const result = parseJsonContent(content);
        if (result.errors.length === 0) {
          currentVariables = result.variables;
        }
      }
      
      if (newFormat === 'env') {
        setContent(variablesToEnvFormat(currentVariables));
      } else {
        setContent(variablesToJsonFormat(currentVariables));
      }
      
      setFormat(newFormat);
    } catch (err) {
      // If conversion fails, just switch format and keep content
      setFormat(newFormat);
    }
  };

  const handleSubmit = () => {
    if (error) return;
    
    onSubmit(previewVariables);
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const fileContent = e.target?.result as string;
      
      // Auto-detect format based on file extension and content
      let detectedFormat: 'env' | 'json' = 'env';
      
      if (file.name.endsWith('.json')) {
        detectedFormat = 'json';
      } else if (file.name.endsWith('.env')) {
        detectedFormat = 'env';
      } else {
        // Use content-based detection
        const detected = detectFormat(fileContent);
        if (detected === 'json') {
          detectedFormat = 'json';
        }
      }
      
      setFormat(detectedFormat);
      setContent(fileContent);
    };
    reader.readAsText(file);
  };

  const variableCount = Object.keys(previewVariables).length;
  const hasChanges = JSON.stringify(previewVariables) !== JSON.stringify(initialVariables);

  return (
    <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50 bg-gradient-to-r from-gray-50/50 to-gray-100/50 dark:from-gray-800/50 dark:to-gray-700/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Code className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Raw Editor
            </h3>
            <span className="inline-flex items-center px-2 py-1 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-xs font-medium text-indigo-700 dark:text-indigo-300">
              {variableCount} variables
            </span>
          </div>
          <button
            type="button"
            onClick={onCancel}
            className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      <div className="p-6 space-y-6">
        {/* Format Selector & File Upload */}
        <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Format:</span>
            <div className="flex bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
              <button
                onClick={() => handleFormatChange('env')}
                className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
                  format === 'env'
                    ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                    : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                }`}
              >
                .env
              </button>
              <button
                onClick={() => handleFormatChange('json')}
                className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
                  format === 'json'
                    ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                    : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                }`}
              >
                JSON
              </button>
            </div>
          </div>
          
          {/* File Upload */}
          <div className="flex items-center">
            <label className="flex items-center px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors cursor-pointer">
              <Upload size={16} className="mr-2" />
              <span className="text-sm font-medium">Upload file</span>
              <input
                type="file"
                accept=".env,.json,.txt"
                onChange={handleFileUpload}
                className="hidden"
              />
            </label>
          </div>
        </div>

        {/* Info Box */}
        <div className="bg-blue-50/80 dark:bg-blue-900/20 border border-blue-200/60 dark:border-blue-800/30 rounded-xl p-4 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <Info className="w-5 h-5 text-blue-500 dark:text-blue-400 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-1">
                {format === 'env' ? '.env Format' : 'JSON Format'}
              </h4>
              <p className="text-sm text-blue-700 dark:text-blue-400">
                {format === 'env' 
                  ? 'Enter variables in KEY=value format, one per line. Use quotes for values with spaces.'
                  : 'Enter variables as a JSON object with key-value pairs.'
                }
              </p>
            </div>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-4 backdrop-blur-sm">
            <div className="flex items-start space-x-3">
              <AlertCircle className="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" />
              <div>
                <h4 className="text-sm font-semibold text-red-800 dark:text-red-300 mb-1">Parse Error</h4>
                <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Warnings Display */}
        {warnings.length > 0 && !error && (
          <div className="bg-yellow-50/80 dark:bg-yellow-900/20 border border-yellow-200/60 dark:border-yellow-700/50 rounded-xl p-4 backdrop-blur-sm">
            <div className="flex items-start space-x-3">
              <AlertCircle className="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5 flex-shrink-0" />
              <div>
                <h4 className="text-sm font-semibold text-yellow-800 dark:text-yellow-200 mb-1">Warnings</h4>
                <ul className="text-sm text-yellow-700 dark:text-yellow-300 space-y-1">
                  {warnings.map((warning, index) => (
                    <li key={index}>• {warning}</li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        )}

        {/* Editor */}
        <div>
          <label className="block text-sm font-semibold text-gray-900 dark:text-white mb-2">
            Variables Content
          </label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder={format === 'env' 
              ? `# Enter your environment variables\nAPI_KEY=your-api-key\nDATABASE_URL="postgresql://user:pass@host:5432/db"\nDEBUG=true`
              : `{\n  "API_KEY": "your-api-key",\n  "DATABASE_URL": "postgresql://user:pass@host:5432/db",\n  "DEBUG": "true"\n}`
            }
            rows={12}
            className="w-full px-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 dark:focus:border-indigo-400 transition-all backdrop-blur-sm resize-none font-mono text-sm"
          />
        </div>

        {/* Preview */}
        {!error && variableCount > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
              Preview ({variableCount} variables)
            </h4>
            <div className="bg-gray-50/50 dark:bg-gray-900/50 rounded-lg border border-gray-200 dark:border-gray-700 max-h-40 overflow-y-auto">
              <div className="p-3 space-y-1">
                {Object.entries(previewVariables).map(([key, value]) => (
                  <div key={key} className="flex items-center space-x-3 text-sm">
                    <span className="font-mono font-medium text-gray-900 dark:text-white min-w-0 flex-shrink-0">
                      {key}
                    </span>
                    <span className="text-gray-400">=</span>
                    <span className="font-mono text-gray-600 dark:text-gray-400 truncate">
                      {value.length > 50 ? `${value.substring(0, 50)}...` : value}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Changes indicator */}
        {hasChanges && !error && (
          <div className="bg-yellow-50/80 dark:bg-yellow-900/20 border border-yellow-200/60 dark:border-yellow-700/50 rounded-xl p-4 backdrop-blur-sm">
            <div className="flex items-center space-x-3">
              <AlertCircle className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
              <p className="text-yellow-800 dark:text-yellow-200 font-medium">
                You have unsaved changes that will replace all existing variables
              </p>
            </div>
          </div>
        )}

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
            onClick={handleSubmit}
            disabled={!!error || !hasChanges}
            className="flex-1 sm:flex-none px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 disabled:shadow-none transition-all duration-200 transform hover:scale-105 disabled:hover:scale-100 flex items-center justify-center"
          >
            <Save size={16} className="mr-2" />
            Apply Changes
          </button>
        </div>
      </div>
    </div>
  );
}