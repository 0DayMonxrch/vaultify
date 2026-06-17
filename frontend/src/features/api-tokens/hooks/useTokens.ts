import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tokensApi, type CreateTokenPayload } from '../../../api/endpoints/tokens.api';
import { queryKeys } from '../../../api/queryKeys';

export const useTokens = () => {
  return useQuery({
    queryKey: queryKeys.tokens.all(),
    queryFn: () => tokensApi.list(),
  });
};

export const useCreateToken = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTokenPayload) => tokensApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tokens.all() });
    },
  });
};

export const useRevokeToken = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (tokenId: string) => tokensApi.revoke(tokenId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tokens.all() });
    },
  });
};
