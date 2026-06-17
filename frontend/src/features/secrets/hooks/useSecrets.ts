import { useQuery } from '@tanstack/react-query';
import { secretsApi } from '../../../api/endpoints/secrets.api';
import { queryKeys } from '../../../api/queryKeys';

export const useSecrets = (projectId: string, env: string) => {
  return useQuery({
    queryKey: queryKeys.secrets.list(projectId, env),
    queryFn: () => secretsApi.list(projectId, env),
    enabled: !!projectId,
  });
};
