import { apiClient } from './client';

export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token?: string;
  user?: User;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface CheckEmailRequest {
  email: string;
}

export interface CheckEmailResponse {
  exists: boolean;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token?: string;
  user?: User;
}

export const authApi = {
  checkEmail: async (data: CheckEmailRequest) => {
    return apiClient.post<CheckEmailResponse>('/api/v1/check-email', data);
  },

  login: async (data: LoginRequest) => {
    return apiClient.post<LoginResponse>('/api/v1/login', data);
  },

  register: async (data: RegisterRequest) => {
    return apiClient.post<{ message: string }>('/api/v1/register', data);
  },

  refreshToken: async (data: RefreshTokenRequest) => {
    return apiClient.post<RefreshTokenResponse>('/api/v1/token/refresh', data);
  },

  me: async () => {
    return apiClient.auth.get<User>('/api/v1/auth/me');
  },
};