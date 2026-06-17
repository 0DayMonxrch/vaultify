import { apiClient } from '../client';

export interface User {
  id: string;
  email: string;
}

export interface TokenResponse {
  access_token: string;
  expires_in: number;
}

export const login = async (credentials: Record<string, any>) => {
  const response = await apiClient.post<TokenResponse>('/auth/login', credentials);
  return response.data;
};

export const register = async (credentials: Record<string, any>) => {
  const response = await apiClient.post<User>('/auth/register', credentials);
  return response.data;
};

export const refresh = async () => {
  const response = await apiClient.post<TokenResponse>('/auth/refresh');
  return response.data;
};

export const logout = async (global: boolean = false) => {
  const response = await apiClient.delete(`/auth/logout${global ? '?global=true' : ''}`);
  return response.data;
};

export const getMe = async () => {
  const response = await apiClient.get<User>('/auth/me');
  return response.data;
};
