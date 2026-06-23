import { Cloud, Code, Database, Server, ShieldCheck, Users } from "lucide-react";

const features = [
  {
    name: 'Git-Driven Deployments',
    description: 'Deploy directly from your Git repositories with automatic builds and deployments',
    icon: Code,
  },
  {
    name: 'Managed Services',
    description: 'One-click deployment of databases, caching solutions, and message brokers',
    icon: Database,
  },
  {
    name: 'Multi-tenant Architecture',
    description: 'Securely isolate resources between different projects and users',
    icon: Users,
  },
  {
    name: 'Kubernetes Native',
    description: 'Built on Kubernetes for enterprise-grade reliability and scalability',
    icon: Cloud,
  },
  {
    name: 'Real-time Monitoring',
    description: 'View logs and deployment status in real-time with advanced filtering',
    icon: Server,
  },
  {
    name: 'User Access Management',
    description: 'Secure user authentication and project ownership controls',
    icon: ShieldCheck,
  },
];

export default function Features() {
  return (
    <div id="features" className="bg-white py-24 dark:bg-gray-900 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto max-w-2xl lg:text-center">
          <h2 className="text-base font-semibold leading-7 text-indigo-600 dark:text-indigo-400">Deploy Faster</h2>
          <p className="mt-2 text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
            Everything you need to deploy your applications
          </p>
          <p className="mt-6 text-lg leading-8 text-gray-600 dark:text-gray-300">
            Kubesa simplifies Kubernetes deployments with a powerful yet intuitive interface,
            allowing users to focus on building great applications instead of managing infrastructure.
          </p>
        </div>
        <div className="mx-auto mt-16 max-w-2xl sm:mt-20 lg:mt-24 lg:max-w-none">
          <dl className="grid max-w-xl grid-cols-1 gap-x-8 gap-y-16 lg:max-w-none lg:grid-cols-3">
            {features.map((feature) => (
              <div key={feature.name} className="flex flex-col">
                <dt className="flex items-center gap-x-3 text-base font-semibold leading-7 text-gray-900 dark:text-white">
                  <feature.icon className="h-5 w-5 flex-none text-indigo-600 dark:text-indigo-500" aria-hidden="true" />
                  {feature.name}
                </dt>
                <dd className="mt-4 flex flex-auto flex-col text-base leading-7 text-gray-600 dark:text-gray-300">
                  <p className="flex-auto">{feature.description}</p>
                </dd>
              </div>
            ))}
          </dl>
        </div>
      </div>
    </div>
  );
}
