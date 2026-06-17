import { useQuery } from '@tanstack/react-query';
import { auditApi } from '../../../api/endpoints/audit.api';
import { queryKeys } from '../../../api/queryKeys';

export const useAuditLog = (projectId: string, page: number) => {
  return useQuery({
    queryKey: queryKeys.audit.list(projectId, page),
    queryFn: () => auditApi.list(projectId, page),
    enabled: !!projectId,
    // Keep previous data on screen while fetching the next page for smoother UX
    placeholderData: (previousData) => previousData,
  });
};
