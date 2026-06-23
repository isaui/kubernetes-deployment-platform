import { useState, useEffect, useRef } from 'react';
import { getBuildLogUrl, getRuntimeLogUrl } from '~/actions/deployment-log';
import { getServiceDeployments } from '~/actions/deployment';
import { Deployment } from '~/types/deployment';
import { AlertCircle, PlayCircle, Terminal, ChevronDown, Activity, Loader2, Clock } from 'lucide-react';

// Type for log entry with timestamp and message
type LogEntry = {
  id: number;
  timestamp: string;
  message: string;
};

// Props for the logs tab component
interface LogsTabProps {
  service: {
    id: string;
    name: string;
    status?: string;
  };
}

export default function LogsTab({ service }: LogsTabProps) {
  // State to track active log type
  const [activeLogType, setActiveLogType] = useState<'build' | 'runtime'>('build');
  
  // State for deployment history
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const [loadingDeployments, setLoadingDeployments] = useState(true);
  const [deploymentsError, setDeploymentsError] = useState<string | null>(null);
  
  // State to store logs for each type
  const [buildLogs, setBuildLogs] = useState<LogEntry[]>([]);
  const [runtimeLogs, setRuntimeLogs] = useState<LogEntry[]>([]);
  
  // State to track connection status
  const [buildLogConnected, setBuildLogConnected] = useState(false);
  const [runtimeLogConnected, setRuntimeLogConnected] = useState(false);
  
  // Error states
  const [buildLogError, setBuildLogError] = useState<string | null>(null);
  const [runtimeLogError, setRuntimeLogError] = useState<string | null>(null);

  // Refs for EventSource objects
  const buildLogSourceRef = useRef<EventSource | null>(null);
  const runtimeLogSourceRef = useRef<EventSource | null>(null);
  
  // Ref for log container to implement auto-scroll
  const logContainerRef = useRef<HTMLDivElement>(null);

  // Fetch deployment history on component mount
  useEffect(() => {
    const fetchDeployments = async () => {
      try {
        setLoadingDeployments(true);
        const deploymentsData = await getServiceDeployments(service.id);
        
        // Sort by createdAt descending (newest first)
        const sortedDeployments = deploymentsData.sort((a, b) => 
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
        
        setDeployments(sortedDeployments);
        
        // Auto-select the latest deployment
        if (sortedDeployments.length > 0) {
          setSelectedDeployment(sortedDeployments[0]);
        }
      } catch (error) {
        setDeploymentsError(error instanceof Error ? error.message : 'Failed to fetch deployments');
      } finally {
        setLoadingDeployments(false);
      }
    };

    fetchDeployments();
  }, [service.id]);

  // Connect to the log streams when selected deployment changes
  useEffect(() => {
    if (!selectedDeployment) return;

    // Function to connect to build logs for specific deployment
    const connectToBuildLogs = (deploymentId: string) => {
      if (buildLogSourceRef.current) {
        buildLogSourceRef.current.close();
      }

      setBuildLogs([]); // Clear previous logs
      setBuildLogError(null);
      const source = new EventSource(getBuildLogUrl(deploymentId), { withCredentials: true });
      buildLogSourceRef.current = source;

      source.onopen = () => {
        setBuildLogConnected(true);
      };

      source.onmessage = (event) => {
        try {
          // Try to parse the data as JSON
          const data = JSON.parse(event.data);
          setBuildLogs((logs) => [
            ...logs,
            {
              id: Date.now() + Math.random(),
              timestamp: new Date().toISOString(),
              message: data.message || event.data,
            },
          ]);
        } catch (e) {
          // If it's not JSON, just display the raw message
          setBuildLogs((logs) => [
            ...logs,
            {
              id: Date.now() + Math.random(),
              timestamp: new Date().toISOString(),
              message: event.data,
            },
          ]);
        }
      };

      source.onerror = () => {
        setBuildLogConnected(false);
        setBuildLogError("Connection to build logs was lost. The build might have completed.");
        source.close();
      };

      return source;
    };

    // Function to connect to runtime logs (always use latest deployment)
    const connectToRuntimeLogs = () => {
      if (runtimeLogSourceRef.current) {
        runtimeLogSourceRef.current.close();
      }

      setRuntimeLogs([]); // Clear previous logs
      setRuntimeLogError(null);
      
      // For runtime logs, always use the latest deployment (first in sorted array)
      const latestDeployment = deployments[0];
      if (!latestDeployment) return null;

      const source = new EventSource(getRuntimeLogUrl(latestDeployment.id), { withCredentials: true });
      runtimeLogSourceRef.current = source;

      source.onopen = () => {
        setRuntimeLogConnected(true);
      };

      source.onmessage = (event) => {
        try {
          // Try to parse the data as JSON
          const data = JSON.parse(event.data);
          setRuntimeLogs((logs) => [
            ...logs,
            {
              id: Date.now() + Math.random(),
              timestamp: new Date().toISOString(),
              message: data.message || event.data,
            },
          ]);
        } catch (e) {
          // If it's not JSON, just display the raw message
          setRuntimeLogs((logs) => [
            ...logs,
            {
              id: Date.now() + Math.random(),
              timestamp: new Date().toISOString(),
              message: event.data,
            },
          ]);
        }
      };

      source.onerror = () => {
        setRuntimeLogConnected(false);
        setRuntimeLogError("Connection to runtime logs was lost. The service might not be running.");
        source.close();
      };

      return source;
    };

    // Connect to log streams
    const buildSource = connectToBuildLogs(selectedDeployment.id);
    const runtimeSource = connectToRuntimeLogs();

    // Clean up on unmount or when selectedDeployment changes
    return () => {
      if (buildSource) {
        buildSource.close();
      }
      if (runtimeSource) {
        runtimeSource.close();
      }
    };
  }, [selectedDeployment, deployments]);

  // Auto-scroll effect
  useEffect(() => {
    const logContainer = logContainerRef.current;
    if (logContainer) {
      logContainer.scrollTop = logContainer.scrollHeight;
    }
  }, [buildLogs, runtimeLogs, activeLogType]);

  // Format deployment option for display
  const formatDeploymentOption = (deployment: Deployment) => {
    const date = new Date(deployment.createdAt).toLocaleString();
    const shortSha = deployment.commitSha?.substring(0, 7) || 'unknown';
    const message = deployment.commitMessage || 'No commit message';
    return `${date} - ${shortSha} - ${message.substring(0, 50)}${message.length > 50 ? '...' : ''}`;
  };

  // Determine which logs to display based on active log type
  const displayLogs = activeLogType === 'build' ? buildLogs : runtimeLogs;
  const isConnected = activeLogType === 'build' ? buildLogConnected : runtimeLogConnected;
  const logError = activeLogType === 'build' ? buildLogError : runtimeLogError;

  if (loadingDeployments) {
    return (
      <div className="p-8">
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-12 text-center">
          <div className="flex items-center justify-center space-x-3">
            <Loader2 className="h-8 w-8 text-indigo-600 animate-spin" />
            <span className="text-lg font-medium text-gray-900 dark:text-white">Loading deployment history...</span>
          </div>
        </div>
      </div>
    );
  }

  if (deploymentsError) {
    return (
      <div className="p-8">
        <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-6 backdrop-blur-sm">
          <div className="flex items-start space-x-3">
            <AlertCircle className="w-6 h-6 text-red-500 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="text-sm font-semibold text-red-800 dark:text-red-200 mb-1">Error Loading Deployments</h3>
              <p className="text-sm text-red-700 dark:text-red-300">{deploymentsError}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (deployments.length === 0) {
    return (
      <div className="p-8">
        <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm p-12 text-center">
          <div className="mx-auto w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-2xl flex items-center justify-center mb-4">
            <Terminal className="w-8 h-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">No Deployments Found</h3>
          <p className="text-gray-500 dark:text-gray-400">Deploy your service first to view logs.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-8">


      {/* Log Type Tabs */}
      <div className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-sm overflow-hidden">
        <div className="border-b border-gray-200/50 dark:border-gray-700/50">
          <nav className="flex" aria-label="Log Types">
            <button
              onClick={() => setActiveLogType('build')}
              className={`flex-1 py-4 px-6 text-center border-b-2 font-medium text-sm transition-all duration-200 ${
                activeLogType === 'build'
                  ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400 bg-indigo-50/50 dark:bg-indigo-900/20'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 hover:bg-gray-50/50 dark:hover:bg-gray-700/30'
              }`}
            >
              <div className="flex items-center justify-center space-x-2">
                <PlayCircle className="w-5 h-5" />
                <span>Build Logs</span>
              </div>
            </button>
            <button
              onClick={() => setActiveLogType('runtime')}
              className={`flex-1 py-4 px-6 text-center border-b-2 font-medium text-sm transition-all duration-200 ${
                activeLogType === 'runtime'
                  ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400 bg-indigo-50/50 dark:bg-indigo-900/20'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 hover:bg-gray-50/50 dark:hover:bg-gray-700/30'
              }`}
            >
              <div className="flex items-center justify-center space-x-2">
                <Terminal className="w-5 h-5" />
                <span>Runtime Logs</span>
              </div>
            </button>
          </nav>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Build logs with grid layout */}
          {activeLogType === 'build' && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6" style={{ minHeight: '600px' }}>
              {/* Left side - Deployment list */}
              <div className="lg:col-span-1">
                <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center">
                  <Clock className="w-4 h-4 mr-2" />
                  Deployments ({deployments.length})
                </h4>
                <div className="bg-gray-50/50 dark:bg-gray-900/50 rounded-xl p-4 h-full overflow-y-auto backdrop-blur-sm" style={{ maxHeight: '550px' }}>
                  <div className="space-y-3">
                    {deployments.map((deployment) => (
                      <div 
                        key={deployment.id}
                        onClick={() => setSelectedDeployment(deployment)}
                        className={`p-4 rounded-xl border cursor-pointer transition-all duration-200 ${
                          selectedDeployment?.id === deployment.id
                            ? 'bg-indigo-100/80 border-indigo-300/60 dark:bg-indigo-900/30 dark:border-indigo-600/50 shadow-sm'
                            : 'bg-white/60 border-gray-200/50 hover:bg-white/80 dark:bg-gray-800/60 dark:border-gray-700/50 dark:hover:bg-gray-700/60 hover:shadow-sm'
                        } backdrop-blur-sm`}
                      >
                        <div className="flex flex-col space-y-3">
                          <div className="flex items-center justify-between">
                            <span className={`inline-flex items-center px-2 py-1 rounded-lg text-xs font-medium ${
                              deployment.status === 'success' 
                                ? 'bg-green-100/80 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                                : deployment.status === 'failed'
                                ? 'bg-red-100/80 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                                : 'bg-yellow-100/80 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                            }`}>
                              {deployment.status}
                            </span>
                          </div>
                          <div className="text-sm font-medium text-gray-900 dark:text-gray-100 line-clamp-2">
                            {deployment.commitMessage || 'No commit message'}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 space-y-1">
                            <div className="font-mono">{deployment.commitSha?.substring(0, 7) || 'unknown'}</div>
                            <div>{new Date(deployment.createdAt).toLocaleString()}</div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Right side - Logs display */}
              <div className="lg:col-span-2 flex flex-col space-y-4">
                {/* Connection status */}
                <div className={`px-4 py-3 rounded-xl text-sm flex items-center backdrop-blur-sm ${
                  buildLogConnected
                    ? 'bg-green-100/80 text-green-800 dark:bg-green-900/30 dark:text-green-400 border border-green-200/60 dark:border-green-700/50'
                    : 'bg-yellow-100/80 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 border border-yellow-200/60 dark:border-yellow-700/50'
                }`}>
                  <span className={`inline-block w-2 h-2 rounded-full mr-3 ${buildLogConnected ? 'bg-green-500' : 'bg-yellow-500'}`}></span>
                  {buildLogConnected ? 'Connected to build logs stream' : 'Not connected to build logs stream'}
                </div>

                {/* Error message if present */}
                {buildLogError && (
                  <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-4 backdrop-blur-sm">
                    <div className="flex items-start space-x-3">
                      <AlertCircle className="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" />
                      <p className="text-sm text-red-700 dark:text-red-300">{buildLogError}</p>
                    </div>
                  </div>
                )}

                {/* Logs display */}
                <div
                  ref={logContainerRef}
                  className="bg-gray-900 text-gray-100 rounded-xl p-6 font-mono text-sm overflow-y-auto flex-1 shadow-inner"
                  style={{ maxHeight: '450px', height: '450px' }}
                >
                  {buildLogs.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-gray-500">
                      <div className="text-center space-y-2">
                        <Terminal className="w-8 h-8 mx-auto" />
                        <p>No build logs available yet.</p>
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-1">
                      {buildLogs.map((log) => (
                        <div key={log.id} className="flex">
                          <span className="text-gray-500 mr-3 flex-shrink-0 w-20">
                            [{new Date(log.timestamp).toLocaleTimeString()}]
                          </span>
                          <span className="flex-1 break-words">{log.message}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Runtime logs - traditional layout */}
          {activeLogType === 'runtime' && (
            <div className="space-y-6">
              {/* Runtime logs info */}
              {deployments.length > 0 && (
                <div className="bg-blue-50/80 dark:bg-blue-900/20 border border-blue-200/60 dark:border-blue-800/30 rounded-xl p-4 backdrop-blur-sm">
                  <div className="flex items-start space-x-3">
                    <Terminal className="w-5 h-5 text-blue-500 mt-0.5 flex-shrink-0" />
                    <div>
                      <h4 className="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-1">Latest Deployment</h4>
                      <p className="text-sm text-blue-700 dark:text-blue-400 font-mono">
                        {formatDeploymentOption(deployments[0])}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Connection status indicator */}
              <div className={`px-4 py-3 rounded-xl text-sm flex items-center backdrop-blur-sm ${
                runtimeLogConnected
                  ? 'bg-green-100/80 text-green-800 dark:bg-green-900/30 dark:text-green-400 border border-green-200/60 dark:border-green-700/50'
                  : 'bg-yellow-100/80 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 border border-yellow-200/60 dark:border-yellow-700/50'
              }`}>
                <span className={`inline-block w-2 h-2 rounded-full mr-3 ${runtimeLogConnected ? 'bg-green-500' : 'bg-yellow-500'}`}></span>
                {runtimeLogConnected ? 'Connected to runtime logs stream' : 'Not connected to runtime logs stream'}
              </div>

              {/* Error message if present */}
              {runtimeLogError && (
                <div className="bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 rounded-xl p-4 backdrop-blur-sm">
                  <div className="flex items-start space-x-3">
                    <AlertCircle className="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" />
                    <p className="text-sm text-red-700 dark:text-red-300">{runtimeLogError}</p>
                  </div>
                </div>
              )}

              {/* Logs display */}
              <div
                ref={logContainerRef}
                className="bg-gray-900 text-gray-100 rounded-xl p-6 font-mono text-sm overflow-y-auto shadow-inner"
                style={{ height: '500px', maxHeight: 'calc(100vh - 300px)', minHeight: '400px' }}
              >
                {runtimeLogs.length === 0 ? (
                  <div className="flex items-center justify-center h-full text-gray-500">
                    <div className="text-center space-y-2">
                      <Terminal className="w-8 h-8 mx-auto" />
                      <p>No runtime logs available yet.</p>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-1">
                    {runtimeLogs.map((log) => (
                      <div key={log.id} className="flex">
                        <span className="text-gray-500 mr-3 flex-shrink-0 w-20">
                          [{new Date(log.timestamp).toLocaleTimeString()}]
                        </span>
                        <span className="flex-1 break-words">{log.message}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}