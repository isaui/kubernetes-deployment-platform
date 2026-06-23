import { Link } from "@remix-run/react";
import { GithubIcon, LinkedinIcon, TwitterIcon } from "lucide-react";

export default function Footer() {
  return (
    <footer className="bg-gray-900">
      <div className="mx-auto max-w-7xl px-6 py-12 md:flex md:items-center md:justify-between">
        <div className="flex justify-center space-x-6 md:order-2">
          {/* GitHub */}
          <a href="https://github.com/isacitra" className="text-gray-400 hover:text-gray-300">
            <span className="sr-only">GitHub</span>
            <GithubIcon className="h-6 w-6" aria-hidden="true" />
          </a>
          {/* LinkedIn */}
          <a href="https://linkedin.com/in/isacitra" className="text-gray-400 hover:text-gray-300">
            <span className="sr-only">LinkedIn</span>
            <LinkedinIcon className="h-6 w-6" aria-hidden="true" />
          </a>
          {/* Twitter */}
          <a href="https://twitter.com/isacitra" className="text-gray-400 hover:text-gray-300">
            <span className="sr-only">Twitter</span>
            <TwitterIcon className="h-6 w-6" aria-hidden="true" />
          </a>
        </div>
        <div className="mt-8 md:order-1 md:mt-0">
          <Link to="/" className="flex items-center justify-center md:justify-start">
            <span className="text-xl font-bold text-white">Kubesa</span>
          </Link>
          <p className="mt-2 text-center text-xs leading-5 text-gray-400 md:text-left">
            &copy; 2025 Isa Citra Buana. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
