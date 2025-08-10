# Store Architecture Documentation

## Overview

This application uses **Zustand** for state management with a modular store architecture. The stores are designed to be:
- **Type-safe** with full TypeScript support
- **Performant** with selective subscriptions and shallow comparisons
- **Persistent** with selective state persistence
- **SSR-friendly** with hydration handling

## Store Structure

```
lib/store/
├── types.ts          # Centralized type definitions
├── user-store.ts     # User session and authentication
├── ui-store.ts       # UI state and interactions
├── config-store.ts   # Application configuration
├── hooks.ts          # Utility hooks for stores
├── index.ts          # Central exports
└── README.md         # This file
```

## Available Stores

### 1. User Store (`useUserStore`)
Manages user authentication and session state.

**State:**
- `user`: Current user object
- `isAuthenticated`: Authentication status
- `isLoading`: Loading state for auth operations
- `lastRefreshed`: Last session refresh timestamp

**Actions:**
- `setUser(user)`: Set the current user
- `updateUser(updates)`: Update user properties
- `updatePreferences(preferences)`: Update user preferences
- `refreshSession()`: Refresh the user session
- `logout()`: Logout the user
- `clearSession()`: Clear session data

**Persistence:** User data and authentication state (localStorage)

### 2. UI Store (`useUIStore`)
Manages UI state and interactions.

**State:**
- `keyboardShortcutsOpen`: Keyboard shortcuts popup state
- `userSettingsOpen`: User settings modal state
- `temporaryChat`: Temporary chat configuration
- `voiceChat`: Voice chat configuration
- `loadingStates`: Named loading states
- `theme`: Application theme

**Actions:**
- `toggleKeyboardShortcuts()`: Toggle shortcuts popup
- `toggleUserSettings()`: Toggle settings modal
- `setTemporaryChat(state)`: Update temporary chat
- `setVoiceChat(state)`: Update voice chat
- `setLoading(key, loading)`: Set named loading state
- `setTheme(theme)`: Set application theme
- `resetUI()`: Reset UI to initial state

**Persistence:** Only theme preference (localStorage)

**Note:** Sidebar state is handled by cookies via the existing `useSidebar` hook to avoid hydration issues.

### 3. Config Store (`useConfigStore`)
Manages application configuration and settings.

**State:**
- `apiUrl`: Backend API URL
- `features`: Feature flags
- `chatConfig`: Chat configuration
- `voiceConfig`: Voice configuration
- `keyboardShortcuts`: Shortcuts configuration
- `locale`: Application locale
- `developerMode`: Developer mode flag

**Actions:**
- `setApiUrl(url)`: Set API URL
- `setFeatures(features)`: Update feature flags
- `setChatConfig(config)`: Update chat config
- `setVoiceConfig(config)`: Update voice config
- `setKeyboardShortcuts(shortcuts)`: Update shortcuts
- `setLocale(locale)`: Set locale
- `toggleDeveloperMode()`: Toggle developer mode
- `loadConfig()`: Load config from server
- `saveConfig()`: Save config to server
- `resetConfig()`: Reset to defaults

**Persistence:** All configuration (localStorage)

## Usage Examples

### Basic Usage
```typescript
import { useUserStore, useUIStore, useConfigStore } from "@/lib/store";

function MyComponent() {
  // Access state
  const user = useUserStore((state) => state.user);
  const theme = useUIStore((state) => state.theme);
  
  // Access actions
  const logout = useUserStore((state) => state.logout);
  const toggleKeyboardShortcuts = useUIStore((state) => state.toggleKeyboardShortcuts);
  
  // Use the state and actions
  return (
    <div>
      <p>Welcome, {user?.name}!</p>
      <button onClick={logout}>Logout</button>
      <button onClick={toggleKeyboardShortcuts}>Show Shortcuts</button>
    </div>
  );
}
```

### Using Selectors (Better Performance)
```typescript
import { useUserStore, userStoreSelectors } from "@/lib/store";

function MyComponent() {
  // Use predefined selectors for better performance
  const userName = useUserStore(userStoreSelectors.userName);
  const isAuthenticated = useUserStore(userStoreSelectors.isAuthenticated);
  
  return <p>{isAuthenticated ? `Hello, ${userName}` : "Please login"}</p>;
}
```

### Multiple Values with Shallow Comparison
```typescript
import { useUIStore } from "@/lib/store";
import { useShallow } from "zustand/shallow";

function MyComponent() {
  // Select multiple values with shallow comparison
  const { theme, temporaryChat } = useUIStore(
    useShallow((state) => ({
      theme: state.theme,
      temporaryChat: state.temporaryChat,
    }))
  );
  
  return (
    <div className={theme}>
      {temporaryChat.isOpen && <TemporaryChat />}
    </div>
  );
}
```

### Handling SSR Hydration
```typescript
import { useStoreHydration } from "@/lib/store/hooks";
import { useUserStore } from "@/lib/store";

function MyComponent() {
  const isHydrated = useStoreHydration();
  const user = useUserStore((state) => state.user);
  
  // Wait for hydration before rendering user-specific content
  if (!isHydrated) {
    return <div>Loading...</div>;
  }
  
  return <div>Welcome, {user?.name}!</div>;
}
```

### Complex Actions
```typescript
import { useStores } from "@/lib/store";

function MyComponent() {
  const { user, ui, config } = useStores();
  
  const handleComplexAction = async () => {
    ui.setLoading("save", true);
    
    try {
      // Update multiple stores
      user.updatePreferences({ displayName: "John" });
      config.setChatConfig({ defaultModel: "gpt-4" });
      
      // Save to server
      await config.saveConfig();
    } finally {
      ui.setLoading("save", false);
    }
  };
  
  return (
    <button 
      onClick={handleComplexAction}
      disabled={ui.loadingStates["save"]}
    >
      Save Settings
    </button>
  );
}
```

## Best Practices

### 1. Use Selectors for Single Values
```typescript
// ❌ Bad - Re-renders on any store change
const store = useUserStore();
const name = store.user?.name;

// ✅ Good - Only re-renders when name changes
const name = useUserStore((state) => state.user?.name);
```

### 2. Use Shallow Comparison for Multiple Values
```typescript
// ❌ Bad - Creates new object every render
const state = useUIStore((state) => ({
  theme: state.theme,
  loading: state.loadingStates,
}));

// ✅ Good - Uses shallow comparison
import { useShallow } from "zustand/shallow";
const state = useUIStore(
  useShallow((state) => ({
    theme: state.theme,
    loading: state.loadingStates,
  }))
);
```

### 3. Handle Hydration in SSR
```typescript
// ❌ Bad - Can cause hydration mismatch
const user = useUserStore((state) => state.user);
return <div>{user?.name}</div>;

// ✅ Good - Waits for hydration
const isHydrated = useStoreHydration();
const user = useUserStore((state) => state.user);

if (!isHydrated) return <div>Loading...</div>;
return <div>{user?.name}</div>;
```

### 4. Use Named Loading States
```typescript
// ❌ Bad - Single loading state for everything
const [loading, setLoading] = useState(false);

// ✅ Good - Named loading states
const setLoading = useUIStore((state) => state.setLoading);
setLoading("save-profile", true);
setLoading("fetch-data", true);
```

### 5. Clean Up Subscriptions
```typescript
useEffect(() => {
  const unsubscribe = useUserStore.subscribe(
    (state) => state.user,
    (user) => {
      console.log("User changed:", user);
    }
  );
  
  // ✅ Always clean up
  return unsubscribe;
}, []);
```

## Migration Guide

If you're migrating from the old `appStore`:

### Old Pattern
```typescript
const [openShortcutsPopup, appStoreMutate] = appStore(
  useShallow((state) => [state.openShortcutsPopup, state.mutate])
);

appStoreMutate((prev) => ({
  openShortcutsPopup: !prev.openShortcutsPopup,
}));
```

### New Pattern
```typescript
const keyboardShortcutsOpen = useUIStore((state) => state.keyboardShortcutsOpen);
const toggleKeyboardShortcuts = useUIStore((state) => state.toggleKeyboardShortcuts);

toggleKeyboardShortcuts();
```

## Testing

### Mocking Stores in Tests
```typescript
import { renderHook } from "@testing-library/react";
import { useUserStore } from "@/lib/store";

// Reset store before each test
beforeEach(() => {
  useUserStore.setState({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    lastRefreshed: null,
  });
});

test("user login", () => {
  const { result } = renderHook(() => useUserStore());
  
  // Test initial state
  expect(result.current.isAuthenticated).toBe(false);
  
  // Test action
  act(() => {
    result.current.setUser({
      id: "1",
      name: "John",
      email: "john@example.com",
    });
  });
  
  expect(result.current.isAuthenticated).toBe(true);
  expect(result.current.user?.name).toBe("John");
});
```

## Troubleshooting

### Hydration Mismatch
If you see hydration errors, make sure to:
1. Use `useStoreHydration` hook for SSR components
2. Don't persist UI state that can cause mismatches (like sidebar state)
3. Use cookies for critical UI state that needs to be server-rendered

### State Not Persisting
Check that:
1. The state is included in the `partialize` function
2. localStorage is not blocked by the browser
3. The store name hasn't changed (would reset persisted state)

### Performance Issues
1. Use selectors to subscribe to specific values
2. Use shallow comparison for multiple values
3. Avoid creating new objects/arrays in selectors
4. Use the predefined selectors when available

## Future Enhancements

Planned improvements:
- [ ] Add chat store for message management
- [ ] Add workflow store for workflow state
- [ ] Add agent store for agent management
- [ ] Add devtools integration for debugging
- [ ] Add middleware for logging and debugging
- [ ] Add cross-tab synchronization
- [ ] Add optimistic updates support