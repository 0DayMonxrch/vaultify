import { apiClient } from '../client';

export interface Project {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  role?: string;
}

export interface CreateProjectPayload {
  name: string;
  slug: string;
}

export interface ProjectMember {
  user_id: string;
  email: string;
  role: string;
}

const normalizeProject = (p: any): Project => ({
  id: p.ID || p.id,
  name: p.Name || p.name,
  slug: p.Slug || p.slug,
  created_at: p.CreatedAt || p.created_at,
  role: p.Role || p.role
});

export const projectsApi = {
  list: async () => {
    const response = await apiClient.get<any[]>('/projects');
    return response.data.map(normalizeProject);
  },
  create: async (data: CreateProjectPayload) => {
    const response = await apiClient.post<any>('/projects', data);
    return normalizeProject(response.data);
  },
  get: async (projectId: string) => {
    const response = await apiClient.get<any>(`/projects/${projectId}`);
    return normalizeProject(response.data);
  },
  update: async (projectId: string, data: CreateProjectPayload) => {
    const response = await apiClient.patch<any>(`/projects/${projectId}`, data);
    return normalizeProject(response.data);
  },
  delete: async (projectId: string) => {
    await apiClient.delete(`/projects/${projectId}`);
  },
  getMembers: async (projectId: string) => {
    try {
      const response = await apiClient.get<ProjectMember[]>(`/projects/${projectId}/members`);
      return response.data;
    } catch {
      return []; // graceful fallback if backend endpoint isn't wired yet
    }
  },
  addMember: async (projectId: string, data: { email: string, role: string }) => {
    const response = await apiClient.post(`/projects/${projectId}/members`, data);
    return response.data;
  },
  removeMember: async (projectId: string, userId: string) => {
    await apiClient.delete(`/projects/${projectId}/members/${userId}`);
  }
};
