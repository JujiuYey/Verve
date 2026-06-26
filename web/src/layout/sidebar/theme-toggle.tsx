import { MoonIcon, SunIcon } from "lucide-react";

import { SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/components/ui/sidebar";
import { useAppStore } from "@/stores/app";

export function ThemeToggle() {
  const theme = useAppStore((s) => s.settings.theme);
  const updateSettings = useAppStore((s) => s.updateSettings);
  const isDark = theme === "dark";
  const { state } = useSidebar();

  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        tooltip={isDark ? "切换到浅色" : "切换到深色"}
        onClick={() => updateSettings({ theme: isDark ? "light" : "dark" })}
      >
        {isDark ? <MoonIcon /> : <SunIcon />}
        {state !== "collapsed" && <span>{isDark ? "深色" : "浅色"}</span>}
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}
