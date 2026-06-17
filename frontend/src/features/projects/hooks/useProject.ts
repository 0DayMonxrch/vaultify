import { useQuery, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '../../../api/endpoints/projects.api';
import { queryKeys } from '../../../api/queryKeys';

export const useProject = (projectId: string) => {
  const queryClient = useQueryClient();
  return useQuery({
    queryKey: queryKeys.projects.detail(projectId),
    queryFn: async () => {
      const data = await projectsApi.get(projectId);
      const projectsList = queryClient.getQueryData<any[]>(queryKeys.projects.all());
      const cachedProject = projectsList?.find(p => p.id === projectId);
      const role = cachedProject?.role || data.role;
      return {
        ...data,
        role,
        isOwner: role === 'owner',
      };
    },
    enabled: !!projectId,
  });
};

export const useProjectMembers = (projectId: string) => {
  return useQuery({
    queryKey: [...queryKeys.projects.detail(projectId), 'members'],
    queryFn: () => projectsApi.getMembers(projectId),
    enabled: !!projectId,
  });
};
