import { PanelLeftCloseIcon } from "lucide-react";

import logo from "@/assets/logo.svg";
import { Button } from "@/components/ui/button";
import { SidebarHeader, useSidebar } from "@/components/ui/sidebar";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

export function Logo() {
  const { state, toggleSidebar } = useSidebar();
  const isCollapsed = state === "collapsed";

  return (
    <SidebarHeader>
      <div className="bg-sidebar-accent ring-sidebar-border/60 flex items-center gap-3 rounded-lg p-2 ring-1 shadow-[0_4px_14px_rgb(15_23_42/0.06),0_1px_2px_rgb(15_23_42/0.04)] group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-0 group-data-[collapsible=icon]:shadow-none group-data-[collapsible=icon]:ring-0">
        {isCollapsed ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                onClick={toggleSidebar}
                aria-label="展开侧边栏"
                className="focus-visible:ring-sidebar-ring rounded-md focus-visible:ring-2 focus-visible:outline-none"
              >
                <img src={logo} alt="SAG Wiki" className="size-9 rounded-md object-contain" />
              </button>
            </TooltipTrigger>
            <TooltipContent side="right">展开侧边栏</TooltipContent>
          </Tooltip>
        ) : (
          <img src={logo} alt="SAG Wiki" className="size-9 shrink-0 rounded-md object-contain" />
        )}
        <div className="flex flex-1 items-center gap-2 group-data-[collapsible=icon]:hidden">
          <div className="grid flex-1 text-left leading-tight">
            <span className="truncate text-[15px] font-semibold tracking-tight">SAG Wiki</span>
            <span className="text-muted-foreground truncate text-xs">智能知识运营中枢</span>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="text-muted-foreground hover:text-foreground size-7 shrink-0"
            onClick={toggleSidebar}
            aria-label="折叠侧边栏"
          >
            <PanelLeftCloseIcon className="size-4" />
          </Button>
        </div>
      </div>
    </SidebarHeader>
  );
}
