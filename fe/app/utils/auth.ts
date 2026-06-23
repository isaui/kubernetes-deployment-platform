import { redirect } from "@remix-run/node";
import { getCurrentUser } from "~/actions/auth.server";

// List of public paths that don't require authentication
const PUBLIC_PATHS = [
  "/login",
  "/register",
  "/",
  // Add any other public routes here
];

// Error message types
const ERROR_TYPES = {
  UNAUTHORIZED: "unauthorized",
  FORBIDDEN: "forbidden",
};

/**
 * Check if the current path is public (doesn't require authentication)
 */
export function isPublicPath(pathname: string): boolean {
  // Exact match for public paths
  if (PUBLIC_PATHS.includes(pathname)) return true;
  
  // Root path special case
  if (pathname === "/") return true;
  
  // Static assets and resources
  if (
    pathname.startsWith("/build/") || 
    pathname.startsWith("/assets/") ||
    pathname.startsWith("/resources/") ||
    pathname.includes(".")
  ) return true;
  
  return false;
}

/**
 * Authentication middleware for handling requests
 * Redirects to login if user is not authenticated and tries to access a protected route
 */
export async function checkAuth(request: Request) {
  const url = new URL(request.url);
  const pathname = url.pathname;
  
  // Skip auth check for public paths
  if (isPublicPath(pathname)) {
    return null;
  }
  
  // Check if user is authenticated
  const user = await getCurrentUser(request);
  
  // If not authenticated, redirect to login with error message
  if (!user) {
    // Store the original URL to redirect back after login
    const loginUrl = new URL("/login", url.origin);
    loginUrl.searchParams.set("redirectTo", pathname);
    loginUrl.searchParams.set("error", ERROR_TYPES.UNAUTHORIZED);
    loginUrl.searchParams.set("message", "Please login to access this page");
    
    return redirect(loginUrl.toString());
  }
  
  // Admin route access control
  if (pathname.startsWith("/admin/") && user.role !== "admin") {
    // User is logged in but doesn't have admin privileges
    const homeUrl = new URL("/login", url.origin);
    homeUrl.searchParams.set("error", ERROR_TYPES.FORBIDDEN);
    homeUrl.searchParams.set("message", "You don't have permission to access this page");
    
    return redirect(homeUrl.toString());
  }
  
  // User is authenticated and has appropriate role
  return null;
}
