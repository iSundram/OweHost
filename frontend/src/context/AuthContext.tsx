import { useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import type { User, LoginRequest } from '../types';
import { AuthContext } from './AuthContextDef';
import { authService } from '../api/services';
export type { AuthContextType } from './AuthContextDef';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('access_token'));
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const bootstrapAuth = async () => {
      const storedAccess = localStorage.getItem('access_token');
      const storedRefresh = localStorage.getItem('refresh_token');

      if (!storedAccess || !storedRefresh) {
        setIsLoading(false);
        return;
      }

      setToken(storedAccess);

      try {
        const userData = await authService.me();
        setUser(userData);
        return;
      } catch {
        // Attempt a token refresh if the access token is no longer valid
        try {
          const refreshed = await authService.refresh(storedRefresh);
          localStorage.setItem('access_token', refreshed.access_token);
          localStorage.setItem('refresh_token', refreshed.refresh_token);
          setToken(refreshed.access_token);
          const userData = await authService.me();
          setUser(userData);
          return;
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError);
        }
      } finally {
        setIsLoading(false);
      }

      // If we reach here the refresh failed – clear stored auth
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setToken(null);
      setUser(null);
      setIsLoading(false);
    };

    bootstrapAuth();
  }, []);

  const login = async (credentials: LoginRequest): Promise<void> => {
    setIsLoading(true);
    try {
      const response = await authService.login(credentials);
      
      if (!response.access_token) {
        throw new Error('Invalid response: missing access token');
      }
      
      localStorage.setItem('access_token', response.access_token);
      localStorage.setItem('refresh_token', response.refresh_token);
      
      setToken(response.access_token);
      
      const userData = await authService.me();
      setUser(userData);
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    const refreshToken = localStorage.getItem('refresh_token');
    try {
      await authService.logout(refreshToken);
    } catch (error) {
      console.warn('Logout request failed, clearing local session instead.', error);
    }

    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
