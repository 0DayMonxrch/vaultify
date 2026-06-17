export const queryKeys = {
  projects: {
    all: () => ['projects'] as const,
    detail: (id: string) => ['projects', id] as const,
  },
  secrets: {
    list: (projectId: string, env: string) => ['projects', projectId, 'secrets', { env }] as const,
  },
  audit: {
    list: (projectId: string, page: number) => ['projects', projectId, 'audit', { page }] as const,
  },
  tokens: {
    all: () => ['tokens'] as const,
  },
};
