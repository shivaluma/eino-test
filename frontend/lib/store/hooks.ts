/**
 * Store Hooks - Utility hooks for store management
 */

import { useEffect, useRef, useState } from "react";
import { shallow } from "zustand/shallow";

/**
 * Hook to handle store hydration on client side
 * Prevents hydration mismatch errors in Next.js
 */
export function useStoreHydration() {
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    setIsHydrated(true);
  }, []);

  return isHydrated;
}

/**
 * Hook to subscribe to store changes with cleanup
 * Usage:
 * useStoreSubscription(useUserStore, (state) => {
 *   console.log("User changed:", state.user);
 * });
 */
export function useStoreSubscription<T>(
  useStore: () => T,
  listener: (state: T) => void,
  deps: any[] = []
) {
  const listenerRef = useRef(listener);
  listenerRef.current = listener;

  useEffect(() => {
    const unsubscribe = (useStore as any).subscribe((state: T) => {
      listenerRef.current(state);
    });

    return unsubscribe;
  }, deps);
}

/**
 * Hook for shallow comparison of store state
 * Prevents unnecessary re-renders
 * Usage:
 * const { user, isAuthenticated } = useShallowStore(
 *   useUserStore,
 *   (state) => ({ user: state.user, isAuthenticated: state.isAuthenticated })
 * );
 */
export function useShallowStore<T, U>(
  useStore: (selector: (state: T) => U, equals?: any) => U,
  selector: (state: T) => U
): U {
  return useStore(selector, shallow);
}

/**
 * Hook to persist value to localStorage with SSR support
 */
export function useLocalStorage<T>(
  key: string,
  initialValue: T,
  options?: {
    serialize?: (value: T) => string;
    deserialize?: (value: string) => T;
  }
): [T, (value: T | ((prev: T) => T)) => void, () => void] {
  const serialize = options?.serialize || JSON.stringify;
  const deserialize = options?.deserialize || JSON.parse;

  // State to store our value
  const [storedValue, setStoredValue] = useState<T>(() => {
    if (typeof window === "undefined") {
      return initialValue;
    }

    try {
      const item = window.localStorage.getItem(key);
      return item ? deserialize(item) : initialValue;
    } catch (error) {
      console.error(`Error loading ${key} from localStorage:`, error);
      return initialValue;
    }
  });

  // Return a wrapped version of useState's setter function that persists the new value to localStorage
  const setValue = (value: T | ((prev: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);

      if (typeof window !== "undefined") {
        window.localStorage.setItem(key, serialize(valueToStore));
      }
    } catch (error) {
      console.error(`Error saving ${key} to localStorage:`, error);
    }
  };

  // Remove value from localStorage
  const removeValue = () => {
    try {
      if (typeof window !== "undefined") {
        window.localStorage.removeItem(key);
      }
      setStoredValue(initialValue);
    } catch (error) {
      console.error(`Error removing ${key} from localStorage:`, error);
    }
  };

  return [storedValue, setValue, removeValue];
}

/**
 * Hook to detect and sync store changes across tabs
 */
export function useCrossTabSync<T>(
  key: string,
  onSync: (data: T) => void
) {
  useEffect(() => {
    if (typeof window === "undefined") return;

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === key && e.newValue) {
        try {
          const data = JSON.parse(e.newValue);
          onSync(data);
        } catch (error) {
          console.error("Error syncing across tabs:", error);
        }
      }
    };

    window.addEventListener("storage", handleStorageChange);
    return () => window.removeEventListener("storage", handleStorageChange);
  }, [key, onSync]);
}

/**
 * Hook for debounced store updates
 * Useful for inputs that update store frequently
 */
export function useDebouncedStore<T>(
  value: T,
  delay: number,
  updateStore: (value: T) => void
) {
  const [debouncedValue, setDebouncedValue] = useState(value);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    timeoutRef.current = setTimeout(() => {
      setDebouncedValue(value);
      updateStore(value);
    }, delay);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [value, delay, updateStore]);

  return debouncedValue;
}