export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  username?: string;
  name?: string;
}

export interface User {
  id: string;
  email: string;
  username?: string;
  name?: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  expiresAt: string;
}

export interface ApiResponse<T> {
  status: string;
  message?: string;
  data?: T;
  error?: string;
}
