import type { LoginRequest, RegisterRequest, AuthResponse, User } from '~/types/auth';
import { getEnv } from '~/utils/env.server';

// Get API URL from environment
const env = getEnv();
const API_URL = `${env.API_BASE_URL}/api/v1`;

/**
 * Login function - sends credentials to backend
 * The backend will set the HTTP-only cookie automatically
 */
export async function login(credentials: LoginRequest) {
  const response = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(credentials),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || data.message || "Login failed");
  }
  
  // Create a new response to forward the Set-Cookie header
  const cookieHeader = response.headers.get("Set-Cookie");
  const remixResponse = new Response(JSON.stringify(data.data), {
    headers: {
      "Content-Type": "application/json",
      ...(cookieHeader ? { "Set-Cookie": cookieHeader } : {}),
    },
  });
  
  return { response: remixResponse, user: data.data.user };
}

/**
 * Register function
 */
export async function register(userData: RegisterRequest) {
  const response = await fetch(`${API_URL}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(userData),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || data.message || "Registration failed");
  }
  
  // Create a new response to forward the Set-Cookie header
  const cookieHeader = response.headers.get("Set-Cookie");
  const remixResponse = new Response(JSON.stringify(data), {
    headers: {
      "Content-Type": "application/json",
      ...(cookieHeader ? { "Set-Cookie": cookieHeader } : {}),
    },
  });
  
  return { response: remixResponse, user: data.user };
}

/**
 * Get current authenticated user
 * 
 * @returns User information if logged in
 */
export async function getCurrentUser(request: Request) {
  try {
    const response = await fetch(`${API_URL}/auth/me`, {
      headers: {
        cookie: request.headers.get("Cookie") || "",
      },
      credentials: "include",
    });

    if (!response.ok) return null;
    
    const data = await response.json();
    return data.user as User;
  } catch (error) {
    console.error("Error fetching current user:", error);
    return null;
  }
}
