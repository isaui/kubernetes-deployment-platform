import { useState } from 'react';
import { Eye, EyeOff, Edit, Trash, Copy, Check } from 'lucide-react';

interface EnvironmentVariableItemProps {
  name: string;
  value: string;
  isEditable: boolean;
  onEdit: (name: string, value: string) => void;
  onDelete: (name: string) => void;
}

export default function EnvironmentVariableItem({
  name,
  value,
  isEditable,
  onEdit,
  onDelete,
}: EnvironmentVariableItemProps) {
  const [showValue, setShowValue] = useState(false);
  const [copied, setCopied] = useState(false);

  const toggleShowValue = () => {
    setShowValue(!showValue);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="grid grid-cols-12 gap-4 px-4 py-3 bg-white/50 dark:bg-gray-800/50 hover:bg-white/80 dark:hover:bg-gray-700/50 rounded-lg border border-gray-200/50 dark:border-gray-700/50 transition-all duration-200 group">
      {/* Name */}
      <div className="col-span-4 flex items-center">
        <span className="font-mono text-sm text-gray-900 dark:text-white font-medium">
          {name}
        </span>
      </div>
      
      {/* Value */}
      <div className="col-span-6 flex items-center space-x-2">
        <div className="flex-1 relative overflow-hidden">
          <span className={`font-mono text-sm text-gray-700 dark:text-gray-300 block truncate group-hover:whitespace-normal group-hover:overflow-visible ${
            showValue ? '' : 'filter blur-sm select-none'
          }`}>
            {showValue ? value : value.replace(/./g, '•')}
          </span>
        </div>
        
        {/* Value Actions */}
        <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={toggleShowValue}
            className="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors"
            title={showValue ? 'Hide value' : 'Show value'}
          >
            {showValue ? <EyeOff size={14} /> : <Eye size={14} />}
          </button>
          
          <button
            onClick={handleCopy}
            className="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors"
            title="Copy value"
          >
            {copied ? <Check size={14} className="text-green-500" /> : <Copy size={14} />}
          </button>
        </div>
      </div>
      
      {/* Actions */}
      {isEditable && (
        <div className="col-span-2 flex items-center justify-end space-x-1">
          <button
            onClick={() => onEdit(name, value)}
            className="p-1.5 text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-md transition-colors opacity-0 group-hover:opacity-100"
            title="Edit variable"
          >
            <Edit size={14} />
          </button>
          
          <button
            onClick={() => onDelete(name)}
            className="p-1.5 text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors opacity-0 group-hover:opacity-100"
            title="Delete variable"
          >
            <Trash size={14} />
          </button>
        </div>
      )}
    </div>
  );
}