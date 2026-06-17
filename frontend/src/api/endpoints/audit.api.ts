import { apiClient } from '../client';

export interface AuditEvent {
  id: string;
  project_id: string;
  user_email: string;
  action: string;
  target_key_name: string;
  ip_address: string;
  created_at: string;
}

export interface PaginatedAuditLog {
  data: AuditEvent[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export const auditApi = {
  list: async (projectId: string, page: number) => {
    try {
      const response = await apiClient.get<PaginatedAuditLog>(`/projects/${projectId}/audit`, {
        params: { page }
      });
      return response.data;
    } catch {
      return { data: [], total: 0, page: 1, per_page: 50, total_pages: 1 };
    }
  }
};
