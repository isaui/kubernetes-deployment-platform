import { GithubIcon, LinkedinIcon, TwitterIcon } from "lucide-react";

export default function About() {
  return (
    <div className="relative isolate overflow-hidden bg-white dark:bg-gray-900 py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto grid max-w-2xl grid-cols-1 gap-x-8 gap-y-16 sm:gap-y-20 lg:mx-0 lg:max-w-none lg:grid-cols-2">
          <div className="lg:pr-8 lg:pt-4">
            <div className="lg:max-w-lg">
              <h2 className="text-base font-semibold leading-7 text-indigo-600 dark:text-indigo-400">Portfolio Project</h2>
              <p className="mt-2 text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">Developed by Isa Citra Buana</p>
              <p className="mt-6 text-lg leading-8 text-gray-600 dark:text-gray-300">
                Kubesa is a showcase project created by Isa Citra Buana, demonstrating expertise in cloud-native technologies 
                and modern web application development.
              </p>
              <div className="mt-10 max-w-xl space-y-8 text-base leading-7 text-gray-600 dark:text-gray-300">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-600/10 text-indigo-600 dark:text-indigo-400">
                      <span className="text-base font-semibold">01</span>
                    </span>
                  </div>
                  <div className="ml-4">
                    <p className="font-semibold text-gray-900 dark:text-white">Full-stack Development</p>
                    <p className="mt-1">
                      Engineered with Go backend (using Gin), Remix frontend, and Kubernetes for orchestration, 
                      showcasing modern full-stack development patterns.
                    </p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-600/10 text-indigo-600 dark:text-indigo-400">
                      <span className="text-base font-semibold">02</span>
                    </span>
                  </div>
                  <div className="ml-4">
                    <p className="font-semibold text-gray-900 dark:text-white">Cloud-Native Architecture</p>
                    <p className="mt-1">
                      Built with containerization and Kubernetes principles at its core, demonstrating advanced 
                      devops knowledge and infrastructure as code concepts.
                    </p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-600/10 text-indigo-600 dark:text-indigo-400">
                      <span className="text-base font-semibold">03</span>
                    </span>
                  </div>
                  <div className="ml-4">
                    <p className="font-semibold text-gray-900 dark:text-white">Real-time Data Streaming</p>
                    <p className="mt-1">
                      Implements Server-Sent Events (SSE) for real-time log streaming and deployment updates,
                      providing instant feedback to users.
                    </p>
                  </div>
                </div>
              </div>
              <div className="mt-10 flex items-center gap-x-6">
                <a 
                  href="https://github.com/isaui" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="rounded-full p-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                >
                  <GithubIcon className="w-5 h-5" />
                </a>
                <a 
                  href="https://linkedin.com/in/isacitra" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="rounded-full p-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                >
                  <LinkedinIcon className="w-5 h-5" />
                </a>
                <a 
                  href="https://twitter.com/isacitra" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="rounded-full p-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                >
                  <TwitterIcon className="w-5 h-5" />
                </a>
              </div>
            </div>
          </div>
          <div className="relative overflow-hidden rounded-xl bg-gray-900 shadow-xl">
            <div className="absolute inset-0 bg-gradient-to-tr from-indigo-300 via-indigo-600/20 to-indigo-900 opacity-10"></div>
            <img
              src="/author.jpg"
              alt="Isa Citra Buana"
              className="aspect-[2/2] md:aspect-[3/4] w-full object-cover"
            />
            <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-gray-900 to-transparent p-6">
              <h3 className="text-lg font-semibold text-white">Isa Citra Buana</h3>
              <p className="mt-1 text-sm text-gray-300">Full Stack Developer & Cloud Engineer</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
