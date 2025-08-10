/**
 * Configuration Store - Manages application configuration and settings
 */

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";
import type { ConfigStore, ConfigState } from "./types";

const initialState: ConfigState = {
  // API Configuration
  apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  
  // Feature Flags
  features: {
    chat: true,
    voice: true,
    agents: true,
    workflows: true,
  },
  
  // Chat Configuration
  chatConfig: {
    defaultModel: "gpt-3.5-turbo",
    streamResponses: true,
    maxMessageLength: 4000,
    enableMarkdown: true,
  },
  
  // Voice Configuration
  voiceConfig: {
    defaultProvider: "openai",
    autoStart: false,
    pushToTalk: false,
  },
  
  // Keyboard Shortcuts
  keyboardShortcuts: {
    enabled: true,
    customShortcuts: {},
  },
  
  // Locale
  locale: "en",
  
  // Developer Mode
  developerMode: false,
};

export const useConfigStore = create<ConfigStore>()(
  persist(
    immer((set, get) => ({
      ...initialState,

      // ========================================================================
      // Configuration Actions
      // ========================================================================

      setApiUrl: (url: string) => {
        set((state) => {
          state.apiUrl = url;
        });
      },

      setFeatures: (features: Partial<ConfigState["features"]>) => {
        set((state) => {
          Object.assign(state.features, features);
        });
      },

      setChatConfig: (config: Partial<ConfigState["chatConfig"]>) => {
        set((state) => {
          Object.assign(state.chatConfig, config);
        });
      },

      setVoiceConfig: (config: Partial<ConfigState["voiceConfig"]>) => {
        set((state) => {
          Object.assign(state.voiceConfig, config);
        });
      },

      setKeyboardShortcuts: (shortcuts: Partial<ConfigState["keyboardShortcuts"]>) => {
        set((state) => {
          Object.assign(state.keyboardShortcuts, shortcuts);
        });
      },

      setLocale: (locale: string) => {
        set((state) => {
          state.locale = locale;
        });
      },

      toggleDeveloperMode: () => {
        set((state) => {
          state.developerMode = !state.developerMode;
        });
      },

      // ========================================================================
      // Load/Save Configuration
      // ========================================================================

      loadConfig: async () => {
        try {
          // Load configuration from API or local storage
          const response = await fetch("/api/config", {
            credentials: "include",
          });
          
          if (response.ok) {
            const config = await response.json();
            set((state) => {
              // Merge remote config with local state
              Object.assign(state, config);
            });
          }
        } catch (error) {
          console.error("Failed to load configuration:", error);
        }
      },

      saveConfig: async () => {
        try {
          const state = get();
          const config = {
            chatConfig: state.chatConfig,
            voiceConfig: state.voiceConfig,
            keyboardShortcuts: state.keyboardShortcuts,
            locale: state.locale,
          };
          
          await fetch("/api/config", {
            method: "PUT",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify(config),
          });
        } catch (error) {
          console.error("Failed to save configuration:", error);
        }
      },

      resetConfig: () => {
        set(() => initialState);
      },
    })),
    {
      name: "config-store",
      storage: createJSONStorage(() => localStorage),
      version: 1,
      migrate: (persistedState: any, version: number) => {
        // Handle migrations when config structure changes
        if (version === 0) {
          // Migration from version 0 to 1
          return {
            ...initialState,
            ...persistedState,
          };
        }
        return persistedState as ConfigState;
      },
    }
  )
);

// ============================================================================
// Selectors - Use these for better performance
// ============================================================================

export const configStoreSelectors = {
  // API
  apiUrl: (state: ConfigStore) => state.apiUrl,
  
  // Features
  isFeatureEnabled: (feature: keyof ConfigState["features"]) => 
    (state: ConfigStore) => state.features[feature],
  allFeatures: (state: ConfigStore) => state.features,
  
  // Chat
  chatConfig: (state: ConfigStore) => state.chatConfig,
  defaultModel: (state: ConfigStore) => state.chatConfig.defaultModel,
  
  // Voice
  voiceConfig: (state: ConfigStore) => state.voiceConfig,
  
  // Shortcuts
  shortcutsEnabled: (state: ConfigStore) => state.keyboardShortcuts.enabled,
  customShortcuts: (state: ConfigStore) => state.keyboardShortcuts.customShortcuts,
  
  // Locale
  locale: (state: ConfigStore) => state.locale,
  
  // Developer
  isDeveloperMode: (state: ConfigStore) => state.developerMode,
};