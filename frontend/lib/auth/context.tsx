'use client';

import React, { createContext, useContext, useCallback, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { authApi } from '@/lib/api/auth';
import { clearUserData } from '@/lib/mutations/auth';
import type { SessionUser } from '@/lib/auth/server';

// Auth context types
interface AuthContextType {
  user: SessionUser | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (user: SessionUser) => void;
  logout: () => Promise<void>;
  refreshUser: (user: SessionUser) => void;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider props
interface AuthProviderProps {
  children: ReactNode;
  initialUser?: SessionUser | null;
}

// Auth provider component
export function AuthProvider({ children, initialUser = null }: AuthProviderProps) {
  const router = useRouter();
  
  // For now, we'll manage user state simply
  // Later this will be enhanced with TanStack Query
  const [user, setUser] = React.useState<SessionUser | null>(initialUser);
  const [isLoading, setIsLoading] = React.useState(false);

  const isAuthenticated = user !== null;

  // Login function - sets user data
  const login = useCallback((userData: SessionUser) => {
    setUser(userData);
  }, []);

  // Refresh user data
  const refreshUser = useCallback((userData: SessionUser) => {
    setUser(userData);
  }, []);

  // Logout function - clears session and redirects
  const logout = useCallback(async () => {
    setIsLoading(true);
    
    try {
      // Call backend to clear cookies
      await authApi.logout();
      
      // Clear local user data
      clearUserData();
      setUser(null);
      
      toast.success('Successfully logged out');
      
      router.push('/sign-in');
      router.refresh();
      
    } catch (error) {
      // Even if API call fails, still clear local data and redirect
      console.error('Logout error:', error);
      clearUserData();
      setUser(null);
      router.push('/sign-in');
      toast.error('Logged out (with errors)');
    } finally {
      setIsLoading(false);
    }
  }, [router]);

  const contextValue: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    refreshUser,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
}

// Hook to use auth context
export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  
  return context;
}

// Helper hook for authentication checks
export function useRequireAuth() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  React.useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/sign-in');
    }
  }, [isAuthenticated, isLoading, router]);

  return { isAuthenticated, isLoading };
}