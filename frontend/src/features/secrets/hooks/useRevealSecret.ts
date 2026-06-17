import { useMutation } from '@tanstack/react-query';
import { secretsApi } from '../../../api/endpoints/secrets.api';

// Note: This mutation must not populate TanStack Query cache with the secret plaintext.
// The data must only flow out of here via the mutation resolution to the component's local state.
export const useRevealSecret = () => {
  return useMutation({
    mutationFn: ({ projectId, secretId }: { projectId: string; secretId: string }) => 
      secretsApi.reveal(projectId, secretId)
  });
};
