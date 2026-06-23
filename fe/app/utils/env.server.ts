/**
 * Server-side environment variables
 */
export function getEnv() {
  return {
    API_BASE_URL: process.env.API_BASE_URL || "http://localhost:8080",
    LOAD_BALANCER_IP: process.env.LOAD_BALANCER_IP || "127.0.0.1",
    NODE_ENV: process.env.NODE_ENV || "production",
  };
}
