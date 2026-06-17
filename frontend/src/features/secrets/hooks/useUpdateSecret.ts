import { useMutation, useQueryClient } from '@tanstack/react-query';
import { secretsApi } from '@/api/endpoints/secrets.api';
import { toast } from 'sonner';

export const useUpdateSecret = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ projectId, secretId, value }: { projectId: string; secretId: string; value: string }) =>
      secretsApi.update(projectId, secretId, value),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['projects', variables.projectId, 'secrets'] });
      toast.success('Secret updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update secret');
    },
  });
};
