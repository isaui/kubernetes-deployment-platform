// Type definitions for window.ENV
declare global {
  interface Window {
    ENV: {
      API_BASE_URL: string;
      LOAD_BALANCER_IP: string;
    };
  }
}

export {};
