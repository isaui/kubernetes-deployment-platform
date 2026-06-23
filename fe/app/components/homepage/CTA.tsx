import { Link } from "@remix-run/react";

export default function CTA() {
  return (
    <div className="bg-indigo-700 dark:bg-indigo-900">
      <div className="mx-auto max-w-7xl px-6 py-24 sm:py-32 lg:px-8 lg:py-40">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Ready to simplify your Kubernetes deployments?
          </h2>
          <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-indigo-200">
            Get started with Kubesa today and experience the power of modern deployment workflows.
            Free for personal use and open-source projects.
          </p>
          <div className="mt-10 flex items-center justify-center gap-x-6">
            <a
              href="/projects"
              className="rounded-md bg-white px-5 py-3 text-base font-semibold text-indigo-700 shadow-sm hover:bg-indigo-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
            >
              Get started
            </a>
            <a href="https://github.com/isaui/kubernetes-deployment-platform" className="text-base font-semibold leading-6 text-white">
              Learn more <span aria-hidden="true">→</span>
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
