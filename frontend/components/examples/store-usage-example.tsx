/**
 * Store Usage Examples - Demonstrates how to use the stores in components
 */

"use client";

import { useEffect } from "react";
import { 
  useUserStore, 
  useUIStore, 
  useConfigStore,
  userStoreSelectors,
  uiStoreSelectors,
  configStoreSelectors,
  useStores,
  useStoreSelector,
  useShallowStore,
  useStoreHydration
} from "@/lib/store";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/**
 * Example 1: Basic Store Usage
 * Direct access to store state and actions
 */
export function BasicStoreExample() {
  // Access entire store
  const userStore = useUserStore();
  const uiStore = useUIStore();
  const configStore = useConfigStore();

  // Or access specific values (better for performance)
  const user = useUserStore((state) => state.user);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const theme = useUIStore((state) => state.theme);

  // Access actions
  const setUser = useUserStore((state) => state.setUser);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Basic Store Usage</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <p>User: {user?.name || "Not logged in"}</p>
          <p>Authenticated: {isAuthenticated ? "Yes" : "No"}</p>
          <p>Theme: {theme}</p>
        </div>
        <div className="space-x-2">
          <Button onClick={toggleSidebar}>Toggle Sidebar</Button>
          <Button onClick={() => setUser(null)}>Logout</Button>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Example 2: Using Selectors for Better Performance
 * Selectors prevent unnecessary re-renders
 */
export function SelectorsExample() {
  // Use predefined selectors
  const userName = useUserStore(userStoreSelectors.userName);
  const preferences = useUserStore(userStoreSelectors.preferences);
  const isDeveloperMode = useConfigStore(configStoreSelectors.isDeveloperMode);
  
  // Use custom selector with useStoreSelector
  const chatConfig = useStoreSelector(useConfigStore, configStoreSelectors.chatConfig);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Using Selectors</CardTitle>
      </CardHeader>
      <CardContent>
        <p>User Name: {userName || "Guest"}</p>
        <p>Display Name: {preferences?.displayName || "Not set"}</p>
        <p>Developer Mode: {isDeveloperMode ? "On" : "Off"}</p>
        <p>Default Model: {chatConfig?.defaultModel}</p>
      </CardContent>
    </Card>
  );
}

/**
 * Example 3: Shallow Comparison for Multiple Values
 * Prevents re-renders when selecting multiple values
 */
export function ShallowComparisonExample() {
  // Without shallow comparison (causes unnecessary re-renders)
  const badExample = useUIStore((state) => ({
    sidebarOpen: state.sidebarOpen,
    theme: state.theme,
  }));

  // With shallow comparison (optimized)
  const goodExample = useShallowStore(
    useUIStore,
    (state) => ({
      sidebarOpen: state.sidebarOpen,
      theme: state.theme,
      keyboardShortcutsOpen: state.keyboardShortcutsOpen,
    })
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Shallow Comparison</CardTitle>
      </CardHeader>
      <CardContent>
        <p>Sidebar: {goodExample.sidebarOpen ? "Open" : "Closed"}</p>
        <p>Theme: {goodExample.theme}</p>
        <p>Shortcuts Popup: {goodExample.keyboardShortcutsOpen ? "Open" : "Closed"}</p>
      </CardContent>
    </Card>
  );
}

/**
 * Example 4: Handling Hydration in SSR
 * Prevents hydration mismatch in Next.js
 */
export function HydrationExample() {
  const isHydrated = useStoreHydration();
  const user = useUserStore((state) => state.user);
  const theme = useUIStore((state) => state.theme);

  // Don't render user-specific content until hydrated
  if (!isHydrated) {
    return (
      <Card>
        <CardContent>
          <p>Loading...</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Hydration Safe Component</CardTitle>
      </CardHeader>
      <CardContent>
        <p>User: {user?.name || "Not logged in"}</p>
        <p>Theme: {theme}</p>
      </CardContent>
    </Card>
  );
}

/**
 * Example 5: Complex Actions and Updates
 * Shows how to handle complex state updates
 */
export function ComplexActionsExample() {
  const { user, ui, config } = useStores();

  const handleComplexAction = async () => {
    // Show loading
    ui.setLoading("complex-action", true);

    try {
      // Update user preferences
      user.updatePreferences({
        displayName: "John Doe",
        profession: "Developer",
      });

      // Update chat config
      config.setChatConfig({
        defaultModel: "gpt-4",
        streamResponses: true,
      });

      // Toggle UI elements
      ui.setTemporaryChat({
        isOpen: true,
        instructions: "Help me with coding",
      });

      // Save config to server
      await config.saveConfig();
    } finally {
      // Clear loading
      ui.setLoading("complex-action", false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Complex Actions</CardTitle>
      </CardHeader>
      <CardContent>
        <Button 
          onClick={handleComplexAction}
          disabled={ui.loadingStates["complex-action"]}
        >
          {ui.loadingStates["complex-action"] ? "Processing..." : "Execute Complex Action"}
        </Button>
      </CardContent>
    </Card>
  );
}

/**
 * Example 6: Subscribing to Store Changes
 * React to store changes outside of React components
 */
export function SubscriptionExample() {
  useEffect(() => {
    // Subscribe to user store changes
    const unsubscribeUser = useUserStore.subscribe(
      (state) => state.user,
      (user) => {
        console.log("User changed:", user);
        // Perform side effects when user changes
        if (user) {
          // User logged in
          console.log("Welcome", user.name);
        } else {
          // User logged out
          console.log("Goodbye");
        }
      }
    );

    // Subscribe to theme changes
    const unsubscribeTheme = useUIStore.subscribe(
      (state) => state.theme,
      (theme) => {
        console.log("Theme changed to:", theme);
        // Apply theme to document
        document.documentElement.className = theme;
      }
    );

    return () => {
      unsubscribeUser();
      unsubscribeTheme();
    };
  }, []);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Store Subscriptions</CardTitle>
      </CardHeader>
      <CardContent>
        <p>Check console for subscription logs</p>
      </CardContent>
    </Card>
  );
}

/**
 * Example 7: Keyboard Shortcuts Integration
 */
export function KeyboardShortcutsExample() {
  const toggleKeyboardShortcuts = useUIStore((state) => state.toggleKeyboardShortcuts);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const setTemporaryChat = useUIStore((state) => state.setTemporaryChat);
  const temporaryChat = useUIStore((state) => state.temporaryChat);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd/Ctrl + K for keyboard shortcuts
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        toggleKeyboardShortcuts();
      }
      
      // Cmd/Ctrl + B for sidebar
      if ((e.metaKey || e.ctrlKey) && e.key === "b") {
        e.preventDefault();
        toggleSidebar();
      }
      
      // Cmd/Ctrl + Shift + T for temporary chat
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === "T") {
        e.preventDefault();
        setTemporaryChat({ 
          ...temporaryChat, 
          isOpen: !temporaryChat.isOpen 
        });
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleKeyboardShortcuts, toggleSidebar, setTemporaryChat, temporaryChat]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Keyboard Shortcuts</CardTitle>
      </CardHeader>
      <CardContent>
        <p>Press Cmd/Ctrl + K to open shortcuts popup</p>
        <p>Press Cmd/Ctrl + B to toggle sidebar</p>
        <p>Press Cmd/Ctrl + Shift + T to toggle temporary chat</p>
      </CardContent>
    </Card>
  );
}

/**
 * Main Example Component - Demonstrates all examples
 */
export default function StoreUsageExamples() {
  return (
    <div className="container mx-auto p-4 space-y-6">
      <h1 className="text-3xl font-bold mb-6">Store Usage Examples</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <BasicStoreExample />
        <SelectorsExample />
        <ShallowComparisonExample />
        <HydrationExample />
        <ComplexActionsExample />
        <SubscriptionExample />
        <KeyboardShortcutsExample />
      </div>
    </div>
  );
}