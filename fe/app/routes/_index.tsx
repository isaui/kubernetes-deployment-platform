import type { MetaFunction } from "@remix-run/node";
import { Hero, Features, About, CTA, Footer } from "~/components/homepage";

export const meta: MetaFunction = () => {
  return [
    { title: "Kubesa - Multi-tenant Kubernetes Deployment Platform" },
    { name: "description", content: "A multi-tenant Kubernetes deployment platform created by Isa Citra Buana as a portfolio showcase." },
  ];
};

export default function Index() {
  return (
    <div className="min-h-screen bg-white dark:bg-gray-900">
      <Hero />
      <Features />
      <About />
      <CTA />
      <Footer />
    </div>
  );
}
