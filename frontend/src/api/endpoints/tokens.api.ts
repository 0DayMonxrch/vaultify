import { apiClient } from '../client';

export interface ApiToken {
  id: string;
  name: string;
  token_prefix: string;
  role: string;
  created_at: string;
  last_used_at: string | null;
  revoked: boolean;
}

export interface CreateTokenPayload {
  project_id: string;
  name: string;
  role: string; // 'read' | 'write'
}

export interface CreateTokenResponse {
  token: string;
}

export const tokensApi = {
  list: async () => {
    const response = await apiClient.get<ApiToken[]>('/tokens');
    return response.data;
  },
  create: async (data: CreateTokenPayload) => {
    const response = await apiClient.post<CreateTokenResponse>('/tokens', data);
    return response.data;
  },
  revoke: async (tokenId: string) => {
    await apiClient.delete(`/tokens/${tokenId}`);
  }
};
