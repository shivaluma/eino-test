'use client';

import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { authApi, type User } from '@/lib/api/auth';
import { useAuth } from '@/lib/auth/context';
import { toast } from 'sonner';

// Query keys for consistent cache management
export const authQueryKeys = {
  session: ['auth', 'session'] as const,
  user: ['auth', 'user'] as const,
} as const;

// Session validation query
export function useSessionQuery() {
  const { logout, refreshUser } = useAuth();

  const query = useQuery({
    queryKey: authQueryKeys.session,
    queryFn: async (): Promise<User> => {
      const response = await authApi.me();
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      if (!response.data) {
        throw new Error('No user data received');
      }
      
      return response.data;
    },
    staleTime: 4 * 60 * 1000, // 4 minutes - refresh before token expiry
    refetchInterval: 5 * 60 * 1000, // Check every 5 minutes
    refetchIntervalInBackground: true,
    retry: (failureCount, error) => {
      // Don't retry if it's an auth error
      if (error?.message?.includes('401') || error?.message?.includes('Unauthenticated')) {
        return false;
      }
      // Retry up to 2 times for other errors
      return failureCount < 2;
    },
  });

  // Handle success and error cases with useEffect
  React.useEffect(() => {
    if (query.data) {
      // Update user in auth context
      const sessionUser = {
        id: query.data.id,
        name: query.data.username,
        email: query.data.email,
        image: null,
      };
      refreshUser(sessionUser);
    }
  }, [query.data, refreshUser]);

  React.useEffect(() => {
    if (query.error) {
      console.error('Session validation failed:', query.error);
      // If session validation fails, logout the user
      logout();
    }
  }, [query.error, logout]);

  return query;
}

// Token refresh mutation
export function useTokenRefresh() {
  const queryClient = useQueryClient();
  const { logout } = useAuth();

  return useMutation({
    mutationFn: async () => {
      const response = await authApi.refreshToken();
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      return response.data;
    },
    onSuccess: () => {
      // Invalidate and refetch session after successful token refresh
      queryClient.invalidateQueries({ queryKey: authQueryKeys.session });
      console.log('Token refreshed successfully');
    },
    onError: (error) => {
      console.error('Token refresh failed:', error);
      // If token refresh fails, logout the user
      logout();
      toast.error('Session expired. Please log in again.');
    },
  });
}

// Hook for proactive token refresh (called before expiry)
export function useProactiveTokenRefresh() {
  const tokenRefresh = useTokenRefresh();
  
  // Function to trigger proactive refresh
  const refreshToken = () => {
    if (!tokenRefresh.isPending) {
      console.log('Proactively refreshing token...');
      tokenRefresh.mutate();
    }
  };

  return {
    refreshToken,
    isRefreshing: tokenRefresh.isPending,
    error: tokenRefresh.error,
  };
}

// Session management utilities
export function useSessionUtils() {
  const queryClient = useQueryClient();
  
  const invalidateSession = () => {
    queryClient.invalidateQueries({ queryKey: authQueryKeys.session });
  };
  
  const clearSessionCache = () => {
    queryClient.removeQueries({ queryKey: authQueryKeys.session });
    queryClient.removeQueries({ queryKey: authQueryKeys.user });
  };

  return {
    invalidateSession,
    clearSessionCache,
  };
}

// Hook that combines session validation with automatic refresh
export function useAuthSession() {
  const sessionQuery = useSessionQuery();
  const tokenRefresh = useProactiveTokenRefresh();
  
  // Auto-refresh token 2 minutes before it would expire
  // Assuming 15-minute access token lifetime
  React.useEffect(() => {
    if (sessionQuery.data && !sessionQuery.error) {
      const refreshInterval = setInterval(() => {
        console.log('Auto-refreshing token...');
        tokenRefresh.refreshToken();
      }, 13 * 60 * 1000); // 13 minutes
      
      return () => clearInterval(refreshInterval);
    }
  }, [sessionQuery.data, sessionQuery.error, tokenRefresh.refreshToken]);

  return {
    user: sessionQuery.data,
    isLoading: sessionQuery.isLoading,
    isError: sessionQuery.isError,
    error: sessionQuery.error,
    isRefreshing: tokenRefresh.isRefreshing,
  };
}