import * as React from "react";

import { Sidebar, SidebarFooter, SidebarMenu, useSidebar } from "@/components/ui/sidebar";

import { Logo } from "./logo";
import { NavSystem } from "./nav-system";
import { SideToggle } from "./side-toggle";
import { ThemeToggle } from "./theme-toggle";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { state, toggleSidebar } = useSidebar();

  return (
    <Sidebar variant="floating" collapsible="icon" {...props}>
      <Logo />

      <NavSystem />

      <SidebarFooter className="mt-auto">
        <SidebarMenu>
          <ThemeToggle />
        </SidebarMenu>
        {state === "collapsed" && (
          <SidebarMenu>
            <SideToggle toggleSidebar={toggleSidebar} />
          </SidebarMenu>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}
