import {
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useLoaderData,
} from "@remix-run/react";
import type { LinksFunction, LoaderFunctionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { getEnv } from "~/utils/env.server";

import "./tailwind.css";

export const links: LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
  // Favicon for light mode
  {
    rel: "icon",
    href: "/logo-dark.png",
    type: "image/png",
    media: "(prefers-color-scheme: light)",
  },
  // Favicon for dark mode
  {
    rel: "icon",
    href: "/logo-light.png",
    type: "image/png",
    media: "(prefers-color-scheme: dark)",
  },
  // Default favicon (for browsers that don't support media queries)
  {
    rel: "icon",
    href: "/logo-dark.png",
    type: "image/png",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export async function loader() {
  // Expose environment variables to the client
  return json({
    ENV: {
      API_BASE_URL: getEnv().API_BASE_URL,
      LOAD_BALANCER_IP: getEnv().LOAD_BALANCER_IP,
    },
  });
}

export default function App() {
  const data = useLoaderData<typeof loader>();
  
  // Make environment variables available to client-side code via window.ENV
  return (
    <>
      <script
        dangerouslySetInnerHTML={{
          __html: `window.ENV = ${JSON.stringify(data.ENV)}`,
        }}
      />
      <Outlet />
    </>
  );
}
