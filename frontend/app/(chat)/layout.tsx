import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/layouts/app-sidebar";
import { AppHeader } from "@/components/layouts/app-header";
import { cookies } from "next/headers";

import { auth } from "@/lib/auth/server";
import { COOKIE_KEY_SIDEBAR_STATE } from "@/lib/const";
import { StoreProvider } from "@/lib/store/store-provider";

// import { AppPopupProvider } from "@/components/layouts/app-popup-provider";

export const experimental_ppr = true;

export default async function ChatLayout({
  children,
}: { children: React.ReactNode }) {
  const cookieStore = await cookies();
  const session = await auth.api
    .getSession()
    .catch(() => null);
  const sidebarCookie = cookieStore.get(COOKIE_KEY_SIDEBAR_STATE)?.value;
  // Default to open if no cookie exists, otherwise use cookie value
  const isCollapsed = sidebarCookie === undefined ? false : sidebarCookie !== "true";
  return (
    <StoreProvider initialSession={session}>
      <SidebarProvider defaultOpen={!isCollapsed}>
        {/* <AppPopupProvider /> */}
        <AppSidebar />
        <main className="relative bg-background w-full flex flex-col h-screen">
          <AppHeader />
          <div className="flex-1 overflow-y-auto">{children}</div>
        </main>
      </SidebarProvider>
    </StoreProvider>
  );
}
