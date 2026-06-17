import axios from 'axios';
import { getToken } from './tokenStore';
import { setupResponseInterceptor } from './interceptors';

export const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

apiClient.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

setupResponseInterceptor(apiClient);

export default apiClient;
