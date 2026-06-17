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
    const response = await apiClient.get<any[]>(`/projects/${projectId}/secrets`, {
      params: { env }
    });
    return response.data.map((s: any) => ({
      id: s.ID,
      project_id: projectId,
      key_name: s.KeyName,
      environment: s.Environment,
      updated_at: s.UpdatedAt,
    })) as Secret[];
  },
  reveal: async (projectId: string, secretId: string) => {
    const response = await apiClient.get<{ value: string }>(`/projects/${projectId}/secrets/${secretId}`);
    return response.data.value;
  },
  delete: async (projectId: string, secretId: string) => {
    await apiClient.delete(`/projects/${projectId}/secrets/${secretId}`);
  },
  update: async (projectId: string, secretId: string, value: string) => {
    const response = await apiClient.put(`/projects/${projectId}/secrets/${secretId}`, { value });
    return response.data;
  }
};
