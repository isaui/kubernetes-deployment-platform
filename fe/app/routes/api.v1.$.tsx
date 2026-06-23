// app/routes/api.v1.$.tsx
import type { ActionFunctionArgs, LoaderFunctionArgs } from "@remix-run/node";
import { getEnv } from "~/utils/env.server";

export async function loader({ request, params }: LoaderFunctionArgs) {
  const env = getEnv();
  process.env.BACKEND_URL = env.API_BASE_URL;
  return proxyToBackend(request, params);
}

export async function action({ request, params }: ActionFunctionArgs) {
  const env = getEnv();
  process.env.BACKEND_URL = env.API_BASE_URL;
  return proxyToBackend(request, params);
}

async function proxyToBackend(request: Request, params: any) {
    const path = params["*"] || "";
    const url = new URL(request.url);
    const backendUrl = `${process.env.BACKEND_URL}/api/v1/${path}${url.search}`;
    
    const hasBody = request.method !== 'GET' && request.method !== 'HEAD';
    
    const response = await fetch(backendUrl, {
      method: request.method,
      headers: Object.fromEntries(request.headers.entries()),
      ...(hasBody && { 
        body: await request.arrayBuffer(),
        duplex: 'half' 
      }),
    });
  
    return new Response(response.body, {
      status: response.status,
      headers: response.headers,
    });
  }