/**
 * Unified App Store - Single store for all application state
 * Simplified architecture with SSR support for Next.js App Router
 */

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";
import type { User, UserPreferences } from "@/types/user";

// ============================================================================
// Store Types
// ============================================================================

export interface AppState {
  // User Session
  user: User | null;
  isAuthenticated: boolean;
  
  // UI State
  keyboardShortcutsOpen: boolean;
  userSettingsOpen: boolean;
  temporaryChat: {
    isOpen: boolean;
    instructions: string;
    model?: string;
  };
  voiceChat: {
    isOpen: boolean;
    agentId?: string;
    provider: string;
    options?: Record<string, any>;
  };
  theme: "light" | "dark" | "system";
  
  // Loading States
  loadingStates: Record<string, boolean>;
  
  // Config
  apiUrl: string;
  locale: string;
  chatConfig: {
    defaultModel?: string;
    streamResponses: boolean;
    maxMessageLength: number;
    enableMarkdown: boolean;
  };
}

export interface AppActions {
  // User Actions
  setUser: (user: User | null) => void;
  updateUser: (updates: Partial<User>) => void;
  updatePreferences: (preferences: Partial<UserPreferences>) => void;
  logout: () => Promise<void>;
  
  // UI Actions
  toggleKeyboardShortcuts: () => void;
  toggleUserSettings: () => void;
  setTemporaryChat: (state: Partial<AppState["temporaryChat"]>) => void;
  setVoiceChat: (state: Partial<AppState["voiceChat"]>) => void;
  setTheme: (theme: AppState["theme"]) => void;
  
  // Loading Actions
  setLoading: (key: string, loading: boolean) => void;
  
  // Config Actions
  setChatConfig: (config: Partial<AppState["chatConfig"]>) => void;
  setLocale: (locale: string) => void;
  
  // Reset Actions
  reset: () => void;
}

export type AppStore = AppState & AppActions;

// ============================================================================
// Initial State
// ============================================================================

const initialState: AppState = {
  // User
  user: null,
  isAuthenticated: false,
  
  // UI
  keyboardShortcutsOpen: false,
  userSettingsOpen: false,
  temporaryChat: {
    isOpen: false,
    instructions: "",
    model: undefined,
  },
  voiceChat: {
    isOpen: false,
    agentId: undefined,
    provider: "openai",
    options: {},
  },
  theme: "system",
  
  // Loading
  loadingStates: {},
  
  // Config
  apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  locale: "en",
  chatConfig: {
    defaultModel: "gpt-3.5-turbo",
    streamResponses: true,
    maxMessageLength: 4000,
    enableMarkdown: true,
  },
};

// ============================================================================
// Create Store (without provider - for client components)
// ============================================================================

const createAppStore = (initState: Partial<AppState> = {}) => {
  return create<AppStore>()(
    persist(
      immer((set, get) => ({
        ...initialState,
        ...initState,

        // ======================================================================
        // User Actions
        // ======================================================================

        setUser: (user: User | null) => {
          set((state) => {
            state.user = user;
            state.isAuthenticated = !!user;
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

        logout: async () => {
          set((state) => {
            state.loadingStates.logout = true;
          });

          try {
            // Use the authApi to call Go backend logout endpoint
            const { authApi } = await import('../api/auth');
            const result = await authApi.logout();
            
            if (result.error) {
              console.error("Logout API error:", result.error);
            }
          } catch (error) {
            console.error("Logout error:", error);
          } finally {
            set((state) => {
              state.user = null;
              state.isAuthenticated = false;
              delete state.loadingStates.logout;
            });
            
            // Redirect to sign-in page
            if (typeof window !== "undefined") {
              window.location.href = "/sign-in";
            }
          }
        },

        // ======================================================================
        // UI Actions
        // ======================================================================

        toggleKeyboardShortcuts: () => {
          set((state) => {
            state.keyboardShortcutsOpen = !state.keyboardShortcutsOpen;
          });
        },

        toggleUserSettings: () => {
          set((state) => {
            state.userSettingsOpen = !state.userSettingsOpen;
          });
        },

        setTemporaryChat: (chatState: Partial<AppState["temporaryChat"]>) => {
          set((state) => {
            Object.assign(state.temporaryChat, chatState);
          });
        },

        setVoiceChat: (voiceState: Partial<AppState["voiceChat"]>) => {
          set((state) => {
            Object.assign(state.voiceChat, voiceState);
          });
        },

        setTheme: (theme: AppState["theme"]) => {
          set((state) => {
            state.theme = theme;
          });
        },

        // ======================================================================
        // Loading Actions
        // ======================================================================

        setLoading: (key: string, loading: boolean) => {
          set((state) => {
            if (loading) {
              state.loadingStates[key] = true;
            } else {
              delete state.loadingStates[key];
            }
          });
        },

        // ======================================================================
        // Config Actions
        // ======================================================================

        setChatConfig: (config: Partial<AppState["chatConfig"]>) => {
          set((state) => {
            Object.assign(state.chatConfig, config);
          });
        },

        setLocale: (locale: string) => {
          set((state) => {
            state.locale = locale;
          });
        },

        // ======================================================================
        // Reset Action
        // ======================================================================

        reset: () => {
          set(() => ({
            ...initialState,
            // Preserve some settings
            theme: get().theme,
            locale: get().locale,
          }));
        },
      })),
      {
        name: "app-store",
        storage: createJSONStorage(() => localStorage),
        partialize: (state) => ({
          // Only persist user preferences and settings
          theme: state.theme,
          locale: state.locale,
          chatConfig: state.chatConfig,
          // Don't persist user session (handled by server)
          // Don't persist UI states or loading states
        }),
      }
    )
  );
};

// ============================================================================
// Store Instance (for client-only components)
// ============================================================================

export const useAppStore = createAppStore();

// ============================================================================
// Store with Initial State (for SSR)
// ============================================================================

export const createAppStoreWithInitialState = (initialState: Partial<AppState>) => {
  return createAppStore(initialState);
};