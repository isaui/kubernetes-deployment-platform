import { Link } from "@remix-run/react";
import { useState, useEffect, useCallback, useRef } from "react";

export default function Hero() {
  const carouselImages = [
    { src: "/project-page.png", alt: "Project management page" },
    { src: "/project-page-2.png", alt: "Project management page" },
    { src: "/service-page.png", alt: "Service details page" },
  ];
  
  const [currentImageIndex, setCurrentImageIndex] = useState(0);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const autoplayTimerRef = useRef<NodeJS.Timeout | null>(null);

  const resetAutoplayTimer = useCallback(() => {
    if (autoplayTimerRef.current) {
      clearInterval(autoplayTimerRef.current);
    }
    
    autoplayTimerRef.current = setInterval(() => {
      setIsTransitioning(true);
      setTimeout(() => {
        setCurrentImageIndex((prevIndex) =>
          prevIndex === carouselImages.length - 1 ? 0 : prevIndex + 1
        );
        setIsTransitioning(false);
      }, 500);
    }, 5000);
  }, [carouselImages.length]);

  const nextImage = useCallback(() => {
    setIsTransitioning(true);
    setTimeout(() => {
      setCurrentImageIndex((prevIndex) =>
        prevIndex === carouselImages.length - 1 ? 0 : prevIndex + 1
      );
      setIsTransitioning(false);
    }, 500); // Match this with the CSS transition duration
    resetAutoplayTimer();
  }, [carouselImages.length, resetAutoplayTimer]);

  const prevImage = useCallback(() => {
    setIsTransitioning(true);
    setTimeout(() => {
      setCurrentImageIndex((prevIndex) =>
        prevIndex === 0 ? carouselImages.length - 1 : prevIndex - 1
      );
      setIsTransitioning(false);
    }, 500); // Match this with the CSS transition duration
    resetAutoplayTimer();
  }, [carouselImages.length, resetAutoplayTimer]);

  useEffect(() => {
    resetAutoplayTimer();
    return () => {
      if (autoplayTimerRef.current) {
        clearInterval(autoplayTimerRef.current);
      }
    };
  }, [resetAutoplayTimer]);

  return (
    <div className="relative isolate overflow-hidden bg-gradient-to-b from-indigo-100/20 dark:from-indigo-900/30">
      <div 
        className="absolute inset-x-0 top-40 -z-10 transform-gpu overflow-hidden blur-3xl" 
        aria-hidden="true"
      >
        <div 
          className="relative left-[calc(50%-11rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 rotate-[30deg] bg-gradient-to-tr from-[#ff80b5] to-[#9089fc] opacity-30 dark:opacity-20 sm:left-[calc(50%-30rem)] sm:w-[72.1875rem]" 
          style={{ clipPath: 'polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)' }}
        />
      </div>
      
      <div className="mx-auto max-w-7xl px-6 pb-24 pt-10 sm:pb-32 lg:flex lg:px-8 lg:py-32">
        <div className="mx-auto max-w-2xl flex-shrink-0 lg:mx-0 lg:max-w-xl">
          <div className="mb-8">
            <span className="rounded-full bg-indigo-500/10 px-3 py-1 text-sm font-semibold leading-6 text-indigo-600 dark:text-indigo-400 ring-1 ring-inset ring-indigo-500/20">
              Open Source
            </span>
          </div>
          <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-6xl">
            Kubesa
          </h1>
          <p className="mt-6 text-xl leading-8 text-gray-600 dark:text-gray-300">
            A multi-tenant Kubernetes deployment platform that simplifies managing containerized applications
          </p>
          <p className="mt-2 text-lg text-gray-600 dark:text-gray-300">
            Deploy, scale, and manage both Git repositories and managed services with ease
          </p>
          <div className="mt-10 flex items-center gap-x-6">
            <a
              href="/projects"
              className="rounded-md bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Get Started
            </a>
            <a href="#features" className="text-sm font-semibold leading-6 text-gray-900 dark:text-gray-300">
              Learn more <span aria-hidden="true">→</span>
            </a>
          </div>
        </div>
        
        {/* Carousel Container */}
        <div className="mx-auto mt-16 flex max-w-2xl sm:mt-24 lg:ml-10 lg:mr-0 lg:mt-0 lg:max-w-none lg:flex-none xl:ml-32 relative">
          <div className="max-w-3xl flex-none sm:max-w-5xl lg:max-w-none overflow-hidden rounded-xl relative">
            {/* Carousel Images */}
            <div className="relative w-[48rem] aspect-[16/9]">
              <div 
                className={`w-full h-full flex items-center justify-center transition-opacity duration-500 ${isTransitioning ? 'opacity-0' : 'opacity-100'}`}
              >
                <img
                  src={carouselImages[currentImageIndex].src}
                  alt={carouselImages[currentImageIndex].alt}
                  className="w-full h-full rounded-md bg-white/5 shadow-2xl ring-1 ring-white/10 object-cover"
                />
              </div>
              
              {/* Navigation Dots */}
              <div className="absolute bottom-4 left-0 right-0 flex justify-center space-x-2">
                {carouselImages.map((_, index) => (
                  <button
                    key={index}
                    className={`h-2 w-10 rounded-full transition-all ${index === currentImageIndex ? 'bg-indigo-600' : 'bg-gray-300/50 dark:bg-gray-600/50'} hover:bg-indigo-500`}
                    onClick={() => {
                      if (index !== currentImageIndex) {
                        setIsTransitioning(true);
                        setTimeout(() => {
                          setCurrentImageIndex(index);
                          setIsTransitioning(false);
                        }, 500);
                        resetAutoplayTimer();
                      }
                    }}
                    aria-label={`Go to slide ${index + 1}`}
                  />
                ))}
              </div>
              
              {/* Navigation Arrows */}
              <button 
                onClick={prevImage}
                className="absolute left-4 top-1/2 -translate-y-1/2 bg-white/30 dark:bg-gray-800/30 backdrop-blur-sm p-2 rounded-full text-gray-800 dark:text-white hover:bg-white/50 dark:hover:bg-gray-800/50 transition-all"
                aria-label="Previous slide"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
                </svg>
              </button>
              <button 
                onClick={nextImage}
                className="absolute right-4 top-1/2 -translate-y-1/2 bg-white/30 dark:bg-gray-800/30 backdrop-blur-sm p-2 rounded-full text-gray-800 dark:text-white hover:bg-white/50 dark:hover:bg-gray-800/50 transition-all"
                aria-label="Next slide"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                </svg>
              </button>
              
              {/* Caption removed as requested */}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
