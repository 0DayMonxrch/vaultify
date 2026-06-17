import { apiClient } from '../client';

export interface Secret {
  id: string;
  project_id: string;
  key_name: string;
  environment: string;
  updated_at: string;
}

export const secretsApi = {
  list: async (projectId: string, env: string) => {
    const response = await apiClient.get<Secret[]>(`/projects/${projectId}/secrets`, {
      params: { env }
    });
    return response.data;
  },
  reveal: async (projectId: string, secretId: string) => {
    const response = await apiClient.get<{ value: string }>(`/projects/${projectId}/secrets/${secretId}`);
    return response.data.value;
  }
};
