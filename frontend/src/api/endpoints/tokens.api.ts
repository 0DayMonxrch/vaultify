import { apiClient } from '../client';

export interface ApiToken {
  id: string;
  name: string;
  prefix: string;
  role_scope: string;
  created_at: string;
  last_used_at: string | null;
}

export interface CreateTokenPayload {
  name: string;
  role_scope: string; // 'read' | 'write'
}

export interface CreateTokenResponse {
  token: ApiToken;
  raw_token: string;
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
