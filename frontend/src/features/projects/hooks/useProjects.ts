import { useQuery } from '@tanstack/react-query';
import { projectsApi } from '../../../api/endpoints/projects.api';
import { queryKeys } from '../../../api/queryKeys';

export const useProjects = () => {
  return useQuery({
    queryKey: queryKeys.projects.all(),
    queryFn: () => projectsApi.list(),
  });
};
