/**
 * UI Store - Manages UI state and interactions
 */

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";
import type { UIStore, UIState } from "./types";

const initialState: UIState = {
  // Sidebar & Navigation
  sidebarOpen: true,
  sidebarWidth: 280,
  
  // Popups & Modals
  keyboardShortcutsOpen: false,
  userSettingsOpen: false,
  
  // Chat UI
  temporaryChat: {
    isOpen: false,
    instructions: "",
    model: undefined,
  },
  
  // Voice Chat
  voiceChat: {
    isOpen: false,
    agentId: undefined,
    provider: "openai",
    options: {},
  },
  
  // Loading States
  loadingStates: {},
  
  // Theme
  theme: "system",
};

export const useUIStore = create<UIStore>()(
  persist(
    immer((set, get) => ({
      ...initialState,

      // ========================================================================
      // Sidebar Actions
      // ========================================================================

      toggleSidebar: () => {
        set((state) => {
          state.sidebarOpen = !state.sidebarOpen;
        });
      },

      setSidebarOpen: (open: boolean) => {
        set((state) => {
          state.sidebarOpen = open;
        });
      },

      setSidebarWidth: (width: number) => {
        set((state) => {
          state.sidebarWidth = Math.max(200, Math.min(400, width));
        });
      },

      // ========================================================================
      // Popup Actions
      // ========================================================================

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

      // ========================================================================
      // Chat Actions
      // ========================================================================

      setTemporaryChat: (chatState: Partial<UIState["temporaryChat"]>) => {
        set((state) => {
          Object.assign(state.temporaryChat, chatState);
        });
      },

      setVoiceChat: (voiceState: Partial<UIState["voiceChat"]>) => {
        set((state) => {
          Object.assign(state.voiceChat, voiceState);
        });
      },

      // ========================================================================
      // Loading Actions
      // ========================================================================

      setLoading: (key: string, loading: boolean) => {
        set((state) => {
          if (loading) {
            state.loadingStates[key] = true;
          } else {
            delete state.loadingStates[key];
          }
        });
      },

      clearLoading: (key: string) => {
        set((state) => {
          delete state.loadingStates[key];
        });
      },

      // ========================================================================
      // Theme Actions
      // ========================================================================

      setTheme: (theme: UIState["theme"]) => {
        set((state) => {
          state.theme = theme;
        });
        
        // Apply theme to document
        if (typeof window !== "undefined") {
          const root = window.document.documentElement;
          root.classList.remove("light", "dark");
          
          if (theme === "system") {
            const systemTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
              ? "dark"
              : "light";
            root.classList.add(systemTheme);
          } else {
            root.classList.add(theme);
          }
        }
      },

      // ========================================================================
      // Reset Action
      // ========================================================================

      resetUI: () => {
        set(() => ({
          ...initialState,
          // Preserve theme preference
          theme: get().theme,
        }));
      },
    })),
    {
      name: "ui-store",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        // Only persist theme preference
        // Sidebar state is handled by cookies to avoid hydration issues
        theme: state.theme,
        // Don't persist popup states, loading states, or sidebar state
      }),
    }
  )
);

// ============================================================================
// Selectors - Use these for better performance
// ============================================================================

export const uiStoreSelectors = {
  // Sidebar
  sidebar: (state: UIStore) => ({
    open: state.sidebarOpen,
    width: state.sidebarWidth,
  }),
  
  // Popups
  popups: (state: UIStore) => ({
    keyboardShortcuts: state.keyboardShortcutsOpen,
    userSettings: state.userSettingsOpen,
  }),
  
  // Chat
  temporaryChat: (state: UIStore) => state.temporaryChat,
  voiceChat: (state: UIStore) => state.voiceChat,
  
  // Loading
  isLoading: (key: string) => (state: UIStore) => state.loadingStates[key] || false,
  anyLoading: (state: UIStore) => Object.keys(state.loadingStates).length > 0,
  
  // Theme
  theme: (state: UIStore) => state.theme,
};