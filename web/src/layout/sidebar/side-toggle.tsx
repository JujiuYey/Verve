import { PanelRightCloseIcon } from "lucide-react";

import { SidebarMenuItem, SidebarMenuButton } from "@/components/ui/sidebar";

export function SideToggle({ toggleSidebar }: { toggleSidebar: () => void }) {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton tooltip="展开侧边栏" onClick={toggleSidebar}>
        <PanelRightCloseIcon />
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}
