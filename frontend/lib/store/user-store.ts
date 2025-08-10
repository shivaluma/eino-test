/**
 * User Store - Manages user session and authentication state
 */

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";
import type { UserStore, UserSession } from "./types";
import type { User, UserPreferences } from "@/types/user";

const initialState: UserSession = {
  user: null,
  isAuthenticated: false,
  isLoading: false,
  lastRefreshed: null,
};

export const useUserStore = create<UserStore>()(
  persist(
    immer((set, get) => ({
      ...initialState,

      // ========================================================================
      // Actions
      // ========================================================================

      setUser: (user: User | null) => {
        set((state) => {
          state.user = user;
          state.isAuthenticated = !!user;
          state.lastRefreshed = user ? new Date() : null;
        });
      },

      updateUser: (updates: Partial<User>) => {
        set((state) => {
          if (state.user) {
            Object.assign(state.user, updates);
          }
        });
      },

      updatePreferences: (preferences: Partial<UserPreferences>) => {
        set((state) => {
          if (state.user) {
            state.user.preferences = {
              ...state.user.preferences,
              ...preferences,
            };
          }
        });
      },

      refreshSession: async () => {
        set((state) => {
          state.isLoading = true;
        });

        try {
          // Call your API to validate and refresh the session
          const response = await fetch("/api/auth/validate", {
            method: "POST",
            credentials: "include",
          });

          if (response.ok) {
            const data = await response.json();
            set((state) => {
              state.user = data.user;
              state.isAuthenticated = true;
              state.lastRefreshed = new Date();
            });
          } else {
            // Session invalid, clear it
            get().clearSession();
          }
        } catch (error) {
          console.error("Failed to refresh session:", error);
          get().clearSession();
        } finally {
          set((state) => {
            state.isLoading = false;
          });
        }
      },

      logout: async () => {
        set((state) => {
          state.isLoading = true;
        });

        try {
          await fetch("/api/auth/logout", {
            method: "POST",
            credentials: "include",
          });
        } catch (error) {
          console.error("Logout error:", error);
        } finally {
          get().clearSession();
        }
      },

      clearSession: () => {
        set(() => initialState);
      },
    })),
    {
      name: "user-store",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        // Only persist user data, not loading states
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        lastRefreshed: state.lastRefreshed,
      }),
      onRehydrateStorage: () => (state) => {
        // After rehydration, check if session is still valid
        if (state?.isAuthenticated && state?.lastRefreshed) {
          const lastRefreshed = new Date(state.lastRefreshed);
          const hoursSinceRefresh = 
            (Date.now() - lastRefreshed.getTime()) / (1000 * 60 * 60);
          
          // Refresh if more than 1 hour old
          if (hoursSinceRefresh > 1) {
            state.refreshSession();
          }
        }
      },
    }
  )
);

// ============================================================================
// Selectors - Use these for better performance
// ============================================================================

export const userStoreSelectors = {
  user: (state: UserStore) => state.user,
  isAuthenticated: (state: UserStore) => state.isAuthenticated,
  isLoading: (state: UserStore) => state.isLoading,
  preferences: (state: UserStore) => state.user?.preferences,
  userId: (state: UserStore) => state.user?.id,
  userName: (state: UserStore) => state.user?.name,
  userEmail: (state: UserStore) => state.user?.email,
};