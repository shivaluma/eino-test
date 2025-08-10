"use client";

/**
 * Store Provider - SSR-compatible store provider for Next.js App Router
 */

import React, { createContext, useContext, useRef, useMemo, useEffect } from "react";
import { useStore } from "zustand";
import { createAppStoreWithInitialState, type AppStore } from "./app-store";
import type { UseBoundStore, StoreApi } from "zustand";
import { authApi } from "@/lib/api/auth";
import { getRefreshState } from "@/lib/api/client";
import type { User } from "@/types/user";

interface StoreProviderProps {
  children: React.ReactNode;
  initialSession?: {
    user: User;
    session: any;
  } | null;
  /**
   * Session validation interval in milliseconds
   * @default 300000 (5 minutes)
   */
  sessionValidationInterval?: number;
  /**
   * Whether to enable passive session validation
   * @default true
   */
  enableSessionValidation?: boolean;
}

// Create store context
type StoreType = ReturnType<typeof createAppStoreWithInitialState>;
const StoreContext = createContext<StoreType | undefined>(undefined);

/**
 * Store Provider component - Initializes store with SSR data
 * Usage: Wrap your app components with this provider
 */
export function StoreProvider({ 
  children, 
  initialSession,
  sessionValidationInterval = 5 * 60 * 1000, // 5 minutes
  enableSessionValidation = true
}: StoreProviderProps) {
  const store = useMemo(() => {
    // Initialize store with server data
    const initialState = initialSession?.user ? {
      user: initialSession.user,
      isAuthenticated: true,
    } : {
      user: null,
      isAuthenticated: false,
    };

    return createAppStoreWithInitialState(initialState);
  }, []);

  return (
    <StoreContext.Provider value={store}>
      {/* Passive session validation */}
      {enableSessionValidation && (
        <SessionValidator interval={sessionValidationInterval} />
      )}
      {children}
    </StoreContext.Provider>
  );
}

/**
 * Hook to use the app store
 * Must be used within StoreProvider
 */
export function useAppStore<T>(selector: (store: AppStore) => T): T {
  const storeContext = useContext(StoreContext);

  if (!storeContext) {
    throw new Error("useAppStore must be used within StoreProvider");
  }

  return useStore(storeContext, selector);
}

/**
 * Hook to get the full store instance
 * Use sparingly - prefer useAppStore with selectors for better performance
 */
export function useAppStoreInstance(): StoreType {
  const storeContext = useContext(StoreContext);

  if (!storeContext) {
    throw new Error("useAppStoreInstance must be used within StoreProvider");
  }

  return storeContext;
}

/**
 * Component to initialize store state from server session
 * Should be placed early in the component tree within StoreProvider
 */
export function StoreInitializer({ 
  initialSession
}: { 
  initialSession?: { user: User; session: any } | null 
}) {
  const setUser = useAppStore((store) => store.setUser);

  React.useEffect(() => {
    // Initialize user from server session on client mount
    if (initialSession?.user) {
      setUser(initialSession.user);
    } else {
      setUser(null);
    }
  }, [initialSession, setUser]);

  return null;
}

/**
 * Session Validator - Passive session validation component
 * Periodically checks session validity and logs out user if needed
 */
export function SessionValidator({ 
  interval = 5 * 60 * 1000 // 5 minutes default
}: { 
  interval?: number 
}) {
  const user = useAppStore((store) => store.user);
  const setUser = useAppStore((store) => store.setUser);
  const logout = useAppStore((store) => store.logout);

  useEffect(() => {
    // Only run validation if user is logged in
    if (!user) return;

    const validateSession = async () => {
      try {
        // Check if token refresh is already in progress
        const { isRefreshing, refreshPromise } = getRefreshState();
        
        if (isRefreshing && refreshPromise) {
          // Wait for ongoing refresh to complete
          console.log('Session validation waiting for ongoing token refresh...');
          await refreshPromise;
          // Refresh completed, now validate
        }
        
        const response = await authApi.me();
        
        if (response.error) {
          // Check if error is due to authentication failure
          const isAuthError = response.error.includes('401') || 
                              response.error.includes('Unauthenticated') ||
                              response.error.includes('Unauthorized');
                              
          if (isAuthError) {
            // Give API client time to handle token refresh and retry
            console.log('Session validation got auth error, waiting for potential refresh...');
            
            // Wait longer to allow any ongoing refresh to complete
            await new Promise(resolve => setTimeout(resolve, 3000)); // Wait 3 seconds
            
            // Check again if refresh is happening
            const { isRefreshing: stillRefreshing } = getRefreshState();
            if (stillRefreshing) {
              console.log('Token refresh still in progress, skipping validation this time');
              return; // Skip this validation cycle
            }
            
            // Retry the validation
            try {
              const retryResponse = await authApi.me();
              if (retryResponse.error) {
                console.log('Session validation failed after retry, logging out user');
                await logout();
              } else if (retryResponse.data) {
                // Success after retry - update user data
                const apiUser = retryResponse.data;
                const sessionUser = {
                  id: apiUser.id,
                  name: apiUser.username,
                  email: apiUser.email,
                  image: null,
                };
                if (JSON.stringify(user) !== JSON.stringify(sessionUser)) {
                  setUser(sessionUser);
                }
              }
            } catch (retryError) {
              console.error('Session validation retry failed:', retryError);
              // Don't logout on network errors during retry
            }
          } else {
            // Non-auth error - log but don't logout
            console.error('Session validation error (non-auth):', response.error);
          }
        } else if (response.data) {
          // Session valid - update user data if it changed
          const apiUser = response.data;
          const sessionUser = {
            id: apiUser.id,
            name: apiUser.username,
            email: apiUser.email,
            image: null, // Adjust based on your API response
          };
          
          // Only update if user data has changed
          if (JSON.stringify(user) !== JSON.stringify(sessionUser)) {
            setUser(sessionUser);
          }
        }
      } catch (error) {
        // Network or other errors - log but don't logout immediately
        // The API client will handle token refresh automatically
        console.error('Session validation error:', error);
      }
    };

    // Initial validation after 30 seconds (give time for app to load)
    const initialTimeout = setTimeout(validateSession, 30 * 1000);

    // Periodic validation based on provided interval
    const validationInterval = setInterval(validateSession, interval);

    return () => {
      clearTimeout(initialTimeout);
      clearInterval(validationInterval);
    };
  }, [user, setUser, logout]);

  return null;
}