import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/layouts/app-sidebar";
import { AppHeader } from "@/components/layouts/app-header";
import { cookies } from "next/headers";

import { auth } from "@/lib/auth/server";
import { COOKIE_KEY_SIDEBAR_STATE } from "@/lib/const";
// import { AppPopupProvider } from "@/components/layouts/app-popup-provider";
import { QueryProvider } from "@/components/layouts/query-provider";

export const experimental_ppr = true;

export default async function ChatLayout({
  children,
}: { children: React.ReactNode }) {
  const cookieStore = await cookies();
  const session = await auth.api
    .getSession()
    .catch(() => null);
  const isCollapsed =
    cookieStore.get(COOKIE_KEY_SIDEBAR_STATE)?.value !== "true";
  return (
    <SidebarProvider defaultOpen={!isCollapsed}>
      <QueryProvider>
        {/* <AppPopupProvider /> */}
        <AppSidebar session={session || undefined} />
        <main className="relative bg-background  w-full flex flex-col h-screen">
          <AppHeader />
          <div className="flex-1 overflow-y-auto">{children}</div>
        </main>
      </QueryProvider>
    </SidebarProvider>
  );
}
