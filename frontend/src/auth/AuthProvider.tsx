import React, { createContext, useState, useEffect, type ReactNode } from 'react';
import { type User, login as apiLogin, logout as apiLogout, refresh as apiRefresh, getMe, demoLogin as apiDemoLogin } from '../api/endpoints/auth.api';
import { setToken, clearToken } from '../api/tokenStore';

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isBootstrapping: boolean;
  login: (credentials: Record<string, any>) => Promise<void>;
  demoLogin: () => Promise<void>;
  logout: (global?: boolean) => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isBootstrapping, setIsBootstrapping] = useState<boolean>(true);

  useEffect(() => {
    const bootstrap = async () => {
      try {
        const { access_token } = await apiRefresh();
        setToken(access_token);
        const me = await getMe();
        setUser(me);
      } catch (err) {
        clearToken();
        setUser(null);
      } finally {
        setIsBootstrapping(false);
      }
    };

    bootstrap();
  }, []);

  const login = async (credentials: Record<string, any>) => {
    const { access_token } = await apiLogin(credentials);
    setToken(access_token);
    const me = await getMe();
    setUser(me);
  };

  const demoLogin = async () => {
    const { access_token } = await apiDemoLogin();
    setToken(access_token);
    const me = await getMe();
    setUser(me);
  };

  const logout = async (global = false) => {
    try {
      await apiLogout(global);
    } catch (e) {
      // Ignored
    } finally {
      clearToken();
      setUser(null);
      window.location.href = '/login';
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isBootstrapping,
        login,
        demoLogin,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
