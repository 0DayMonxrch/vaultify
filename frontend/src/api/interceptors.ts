import type { AxiosInstance, AxiosError } from 'axios';
import { setToken, clearToken } from './tokenStore';

let isRefreshing = false;
let refreshPromise: Promise<string> | null = null;

export const setupResponseInterceptor = (apiClient: AxiosInstance) => {
  apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const originalRequest = error.config;

      if (!originalRequest) {
        return Promise.reject(error);
      }

      const url = originalRequest.url || '';
      const isAuthRoute = url.includes('/auth/login') || 
                          url.includes('/auth/register') || 
                          url.includes('/auth/refresh');

      if (error.response?.status === 401 && !isAuthRoute && !(originalRequest as any)._retry) {
        (originalRequest as any)._retry = true;

        if (!isRefreshing) {
          isRefreshing = true;
          refreshPromise = apiClient.post<{ access_token: string }>('/auth/refresh')
            .then((res) => {
              const newToken = res.data.access_token;
              setToken(newToken);
              return newToken;
            })
            .catch((refreshError) => {
              clearToken();
              window.location.href = '/login';
              return Promise.reject(refreshError);
            })
            .finally(() => {
              isRefreshing = false;
              refreshPromise = null;
            });
        }

        try {
          const newToken = await refreshPromise!;
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return apiClient(originalRequest);
        } catch (refreshErr) {
          // On refresh failure, reject the original failed request without auto-retrying
          return Promise.reject(error);
        }
      }

      return Promise.reject(error);
    }
  );
};
