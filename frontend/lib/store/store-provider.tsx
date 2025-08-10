"use client";

/**
 * Store Provider - SSR-compatible store provider for Next.js App Router
 */

import React, { createContext, useContext, useRef, useMemo } from "react";
import { useStore } from "zustand";
import { createAppStoreWithInitialState, type AppStore } from "./app-store";
import type { UseBoundStore, StoreApi } from "zustand";
import type { User } from "@/types/user";

interface StoreProviderProps {
  children: React.ReactNode;
  initialSession?: {
    user: User;
    session: any;
  } | null;
}

// Create store context
type StoreType = ReturnType<typeof createAppStoreWithInitialState>;
const StoreContext = createContext<StoreType | undefined>(undefined);

/**
 * Store Provider component - Initializes store with SSR data
 * Usage: Wrap your app components with this provider
 */
export function StoreProvider({ children, initialSession }: StoreProviderProps) {
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