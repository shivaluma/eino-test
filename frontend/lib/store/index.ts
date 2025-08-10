/**
 * Store Exports - Central export point for all stores
 */

// Export stores
export { useUserStore, userStoreSelectors } from "./user-store";
export { useUIStore, uiStoreSelectors } from "./ui-store";
export { useConfigStore, configStoreSelectors } from "./config-store";

// Export types
export type {
  // User types
  UserStore,
  UserSession,
  UserActions,
  
  // UI types
  UIStore,
  UIState,
  UIActions,
  
  // Config types
  ConfigStore,
  ConfigState,
  ConfigActions,
  
  // Chat types
  ChatStore,
  ChatState,
  ChatActions,
  ChatMessage,
  ChatThread,
  
  // Utility types
  StoreSelector,
  StoreMutator,
  StoreListener,
  PersistConfig,
} from "./types";

// Export hooks
export { useStoreHydration, useStoreSubscription } from "./hooks";

/**
 * Combined store hook for components that need multiple stores
 * Usage:
 * const { user, ui, config } = useStores();
 */
export function useStores() {
  const user = useUserStore();
  const ui = useUIStore();
  const config = useConfigStore();
  
  return { user, ui, config };
}

/**
 * Typed selector hook for better performance
 * Usage:
 * const userName = useStoreSelector(useUserStore, userStoreSelectors.userName);
 */
export function useStoreSelector<T, U>(
  useStore: () => T,
  selector: (state: T) => U
): U {
  return useStore()(selector as any);
}