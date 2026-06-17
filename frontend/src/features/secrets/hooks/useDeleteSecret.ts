import { useMutation, useQueryClient } from '@tanstack/react-query';
import { secretsApi } from '@/api/endpoints/secrets.api';
import { toast } from 'sonner';

export const useDeleteSecret = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ projectId, secretId }: { projectId: string; secretId: string }) =>
      secretsApi.delete(projectId, secretId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['projects', variables.projectId, 'secrets'] });
      toast.success('Secret deleted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete secret');
    },
  });
};
