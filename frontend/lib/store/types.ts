/**
 * Store Types - Centralized type definitions for all stores
 */

import type { User, UserPreferences } from "@/types/user";

// ============================================================================
// User Store Types
// ============================================================================

export interface UserSession {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  lastRefreshed: Date | null;
}

export interface UserActions {
  setUser: (user: User | null) => void;
  updateUser: (updates: Partial<User>) => void;
  updatePreferences: (preferences: Partial<UserPreferences>) => void;
  refreshSession: () => Promise<void>;
  logout: () => Promise<void>;
  clearSession: () => void;
}

export type UserStore = UserSession & UserActions;

// ============================================================================
// UI Store Types
// ============================================================================

export interface UIState {
  // Sidebar & Navigation
  sidebarOpen: boolean;
  sidebarWidth: number;
  
  // Popups & Modals
  keyboardShortcutsOpen: boolean;
  userSettingsOpen: boolean;
  
  // Chat UI
  temporaryChat: {
    isOpen: boolean;
    instructions: string;
    model?: string;
  };
  
  // Voice Chat
  voiceChat: {
    isOpen: boolean;
    agentId?: string;
    provider: string;
    options?: Record<string, any>;
  };
  
  // Loading States
  loadingStates: Record<string, boolean>;
  
  // Theme
  theme: "light" | "dark" | "system";
}

export interface UIActions {
  // Sidebar
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  setSidebarWidth: (width: number) => void;
  
  // Popups
  toggleKeyboardShortcuts: () => void;
  toggleUserSettings: () => void;
  
  // Chat
  setTemporaryChat: (state: Partial<UIState["temporaryChat"]>) => void;
  setVoiceChat: (state: Partial<UIState["voiceChat"]>) => void;
  
  // Loading
  setLoading: (key: string, loading: boolean) => void;
  clearLoading: (key: string) => void;
  
  // Theme
  setTheme: (theme: UIState["theme"]) => void;
  
  // Reset
  resetUI: () => void;
}

export type UIStore = UIState & UIActions;

// ============================================================================
// Configuration Store Types
// ============================================================================

export interface ConfigState {
  // API Configuration
  apiUrl: string;
  
  // Feature Flags
  features: {
    chat: boolean;
    voice: boolean;
    agents: boolean;
    workflows: boolean;
  };
  
  // Chat Configuration
  chatConfig: {
    defaultModel?: string;
    streamResponses: boolean;
    maxMessageLength: number;
    enableMarkdown: boolean;
  };
  
  // Voice Configuration
  voiceConfig: {
    defaultProvider: string;
    autoStart: boolean;
    pushToTalk: boolean;
  };
  
  // Keyboard Shortcuts
  keyboardShortcuts: {
    enabled: boolean;
    customShortcuts?: Record<string, string>;
  };
  
  // Locale
  locale: string;
  
  // Developer Mode
  developerMode: boolean;
}

export interface ConfigActions {
  setApiUrl: (url: string) => void;
  setFeatures: (features: Partial<ConfigState["features"]>) => void;
  setChatConfig: (config: Partial<ConfigState["chatConfig"]>) => void;
  setVoiceConfig: (config: Partial<ConfigState["voiceConfig"]>) => void;
  setKeyboardShortcuts: (shortcuts: Partial<ConfigState["keyboardShortcuts"]>) => void;
  setLocale: (locale: string) => void;
  toggleDeveloperMode: () => void;
  loadConfig: () => Promise<void>;
  saveConfig: () => Promise<void>;
  resetConfig: () => void;
}

export type ConfigStore = ConfigState & ConfigActions;

// ============================================================================
// Chat Store Types (simplified version for your needs)
// ============================================================================

export interface ChatMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: Date;
  metadata?: Record<string, any>;
}

export interface ChatThread {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: Date;
  updatedAt: Date;
  archived: boolean;
}

export interface ChatState {
  threads: ChatThread[];
  currentThreadId: string | null;
  isGenerating: boolean;
  generatingThreadIds: Set<string>;
}

export interface ChatActions {
  // Thread Management
  createThread: (title?: string) => ChatThread;
  selectThread: (threadId: string | null) => void;
  updateThread: (threadId: string, updates: Partial<ChatThread>) => void;
  deleteThread: (threadId: string) => void;
  archiveThread: (threadId: string) => void;
  
  // Message Management
  addMessage: (threadId: string, message: Omit<ChatMessage, "id" | "timestamp">) => void;
  updateMessage: (threadId: string, messageId: string, content: string) => void;
  deleteMessage: (threadId: string, messageId: string) => void;
  
  // Generation State
  setGenerating: (threadId: string, generating: boolean) => void;
  
  // Bulk Operations
  clearAllThreads: () => void;
  loadThreads: () => Promise<void>;
  saveThreads: () => Promise<void>;
}

export type ChatStore = ChatState & ChatActions;

// ============================================================================
// Root Store Type (combines all stores)
// ============================================================================

export interface RootStore {
  user: UserStore;
  ui: UIStore;
  config: ConfigStore;
  chat: ChatStore;
}

// ============================================================================
// Persistence Types
// ============================================================================

export interface PersistConfig<T> {
  name: string;
  version?: number;
  partialize?: (state: T) => Partial<T>;
  migrate?: (persistedState: any, version: number) => T;
  storage?: "localStorage" | "sessionStorage" | "indexedDB";
  skipHydration?: boolean;
}

// ============================================================================
// Store Utilities
// ============================================================================

export type StoreSelector<T, U> = (state: T) => U;
export type StoreMutator<T> = (state: T) => Partial<T> | void;
export type StoreListener<T> = (state: T, prevState: T) => void;