import * as React from "react";

import { Sidebar, SidebarFooter, SidebarMenu } from "@/components/ui/sidebar";

import { Logo } from "./logo";
import { NavSystem } from "./nav-system";
import { ThemeToggle } from "./theme-toggle";
import { User } from "./user";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <Logo />

      <NavSystem />

      <SidebarFooter className="mt-auto">
        <SidebarMenu>
          <ThemeToggle />
        </SidebarMenu>
        <User />
      </SidebarFooter>
    </Sidebar>
  );
}
