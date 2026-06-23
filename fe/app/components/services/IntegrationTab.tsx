import { useState } from 'react';
import { ClipboardCopy, Check, RefreshCw, Hammer, Github } from 'lucide-react';
import type { Service } from '~/types/service';

interface IntegrationTabProps {
  service: Service;
}

export default function IntegrationTab({ service }: IntegrationTabProps) {
  const [copied, setCopied] = useState<string | null>(null);
  
  // Only show for Git-based services
  if (service.type !== 'git') {
    return (
      <div className="p-6 text-center">
        <div className="py-12">
          <p className="text-gray-500 dark:text-gray-400 text-lg">
            CI/CD integration is not available for managed services.
          </p>
        </div>
      </div>
    );
  }
  
  // Generate API endpoints based on the backend implementation
  const deployEndpoint = `${window.ENV.API_BASE_URL}/api/v1/deployments/git`;
  
  // Example cURL commands for CI/CD integration
  const deployCommand = `curl -X POST ${deployEndpoint} \\
-H "Content-Type: application/json" \\
-d '{
    "serviceId": "${service.id}",
    "apiKey": "${service.apiKey || 'YOUR_SERVICE_API_KEY'}", 
    "commitId": "YOUR_COMMIT_SHA",
    "commitMessage": "Optional commit message"
  }'`;

  // Copy to clipboard
  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopied(id);
    setTimeout(() => setCopied(null), 2000);
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">CI/CD Integration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
          Use these API endpoints to integrate with your CI/CD pipeline. You'll need the service's API key from the Information tab to authenticate requests. Check deployment status and logs in the web interface.
        </p>
        <div className="bg-amber-50 dark:bg-amber-900/30 border border-amber-100 dark:border-amber-800/50 rounded-lg p-3 mt-2">
          <div className="flex">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-amber-500 mt-0.5 mr-2 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div>
              <p className="text-sm font-medium text-amber-800 dark:text-amber-300">Repository Requirements</p>
              <p className="text-xs text-amber-700 dark:text-amber-400">Currently, we only support <strong>public Git repositories</strong> that include a <strong>Dockerfile</strong> at the root of the repository. Private repositories will be supported in a future update.</p>
            </div>
          </div>
        </div>
      </div>
      
      <div className="mb-8 space-y-6">
        {/* Deploy Endpoint */}
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <div className="p-4 bg-gray-50 dark:bg-gray-800">
            <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-2 flex items-center gap-1">
              <Hammer className="size-4" /> Deploy
            </h4>
            <p className="text-xs text-gray-600 dark:text-gray-400 mb-4">
              Trigger a deployment using your service's API key and a Git commit SHA. The system will clone your repository, build your Dockerfile, and deploy the container.
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">Endpoint:</p>
            <div className="flex items-center mb-4">
              <span className="font-mono text-sm bg-gray-100 dark:bg-gray-900 p-2 rounded flex-grow">
                POST {deployEndpoint}
              </span>
              <button 
                onClick={() => handleCopy(`POST ${deployEndpoint}`, 'deploy-endpoint')}
                className="ml-2 p-2 text-gray-500 hover:text-gray-700"
                title="Copy to clipboard"
              >
                {copied === 'deploy-endpoint' ? <Check className="w-4 h-4" /> : <ClipboardCopy className="w-4 h-4" />}
              </button>
            </div>
            
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">Example:</p>
            <div className="relative">
              <pre className="bg-gray-100 dark:bg-gray-900 p-4 rounded-md overflow-auto text-sm">{deployCommand}</pre>
              <button 
                onClick={() => handleCopy(deployCommand, 'deploy-command')}
                className="absolute top-2 right-2 p-2 bg-gray-200 dark:bg-gray-800 rounded"
                title="Copy to clipboard"
              >
                {copied === 'deploy-command' ? <Check className="w-4 h-4" /> : <ClipboardCopy className="w-4 h-4" />}
              </button>
            </div>
          </div>
        </div>

        {/* GitHub Actions */}
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <div className="bg-gray-50 dark:bg-gray-800 px-4 py-3 flex items-center">
            <Github className="w-5 h-5 text-indigo-500 mr-2" />
            <h4 className="font-medium text-gray-900 dark:text-gray-100">GitHub Actions Workflow</h4>
          </div>
          
          <div className="p-4">
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3 flex items-center">
              <span className="font-mono bg-gray-100 dark:bg-gray-900 px-2 py-1 rounded mr-2">.github/workflows/deploy.yml</span>
              <span>Save this file in your repository:</span>
            </p>
            
            {/* GitHub Actions YAML example */}
            <div className="relative">
              <pre className="bg-gray-100 dark:bg-gray-900 p-4 rounded-md overflow-auto text-sm">{`name: Deploy to Kubesa

on:
  push:
    branches: [ ${service.branch || 'main'} ]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Deploy to Kubesa
        run: |
          echo "Triggering deployment..."          
          curl -X POST ${deployEndpoint} \\
            -H "Content-Type: application/json" \\
            -d '{
              "serviceId": "${service.id}",
              "apiKey": "\${{ secrets.KUBESA_API_KEY }}",
              "commitId": "\${{ github.sha }}",
              "commitMessage": "\${{ github.event.head_commit.message }}"
            }'
          
          echo "Deployment initiated. Check deployment status and logs in the Kubesa web interface."`}</pre>
              <button 
                onClick={() => handleCopy(`name: Deploy to Kubesa

on:
  push:
    branches: [ ${service.branch || 'main'} ]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Deploy to Kubesa
        run: |
          echo "Triggering deployment..."          
          curl -X POST ${deployEndpoint} \\
            -H "Content-Type: application/json" \\
            -d '{
              "serviceId": "${service.id}",
              "apiKey": "\${{ secrets.KUBESA_API_KEY }}",
              "commitId": "\${{ github.sha }}",
              "commitMessage": "\${{ github.event.head_commit.message }}"
            }'
          
          echo "Deployment initiated. Check deployment status and logs in the Kubesa web interface."`, 'github-action')}
                className="absolute top-2 right-2 p-2 bg-gray-200 dark:bg-gray-800 rounded"
                title="Copy to clipboard"
              >
                {copied === 'github-action' ? <Check className="w-4 h-4" /> : <ClipboardCopy className="w-4 h-4" />}
              </button>
            </div>
            
            <div className="mt-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-900/50 rounded-md p-3 flex items-start">
              <div className="text-yellow-600 dark:text-yellow-400 mr-2 mt-0.5">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="text-sm text-yellow-800 dark:text-yellow-300">
                <p className="font-medium">Security note</p>
                <p className="mt-1">Store your API key as a GitHub repository secret named <code className="font-mono bg-yellow-100 dark:bg-yellow-900/40 px-1 py-0.5 rounded">KUBESA_API_KEY</code></p>
              </div>
            </div>

            <div className="mt-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-900/50 rounded-md p-3 flex items-start">
              <div className="text-blue-600 dark:text-blue-400 mr-2 mt-0.5">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="text-sm text-blue-800 dark:text-blue-300">
                <p className="font-medium">Monitoring deployments</p>
                <p className="mt-1">After triggering a deployment, visit the Kubesa web interface to monitor deployment progress, view logs, and check the status of your service.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}