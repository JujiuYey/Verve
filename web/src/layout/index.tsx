import { Outlet } from "@tanstack/react-router";

import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

import { AppSidebar } from "./sidebar";

export function LayoutComponent() {
  return (
    <SidebarProvider defaultOpen className="[background-image:none] bg-background">
      <AppSidebar />
      <SidebarInset className="md:py-3 md:pr-3">
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  );
}
