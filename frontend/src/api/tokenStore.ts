let accessToken: string | null = null;

export const getToken = (): string | null => {
  return accessToken;
};

export const setToken = (token: string | null): void => {
  accessToken = token;
};

export const clearToken = (): void => {
  accessToken = null;
};
