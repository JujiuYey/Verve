import { IconDotsVertical, IconLogout, IconSettings, IconUserCircle } from "@tabler/icons-react";
import { Link, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useSidebar } from "@/components/ui/sidebar";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useAuthStore } from "@/stores/auth";

export function User() {
  const { isMobile, state } = useSidebar();
  const isCollapsed = state === "collapsed";
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const clearAuth = useAuthStore((s) => s.clearAuth);

  const name = user?.full_name || user?.username || user?.email || "未登录";
  const subtitle = user?.email || "知识库协作者";
  const avatarUrl = user?.avatar || "";

  const getFallback = () => (name ? name.charAt(0).toUpperCase() : "U");

  const handleLogout = () => {
    clearAuth();
    toast.success("已退出登录");
    navigate({ to: "/login" });
  };

  const avatar = (
    <Avatar className="size-8 rounded-md">
      <AvatarImage src={avatarUrl} alt={name} />
      <AvatarFallback className="rounded-md">{getFallback()}</AvatarFallback>
    </Avatar>
  );

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          aria-label={isCollapsed ? name : undefined}
          className="bg-sidebar-accent ring-sidebar-border/60 hover:bg-sidebar-accent/80 focus-visible:ring-sidebar-ring flex items-center gap-3 rounded-lg p-2 ring-1 shadow-[0_4px_14px_rgb(15_23_42/0.06),0_1px_2px_rgb(15_23_42/0.04)] transition-colors focus-visible:ring-2 focus-visible:outline-none group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-0 group-data-[collapsible=icon]:shadow-none group-data-[collapsible=icon]:ring-0 group-data-[collapsible=icon]:hover:bg-transparent data-[state=open]:bg-sidebar-accent/80"
        >
          {isCollapsed ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="flex">{avatar}</span>
              </TooltipTrigger>
              <TooltipContent side="right">{name}</TooltipContent>
            </Tooltip>
          ) : (
            avatar
          )}
          <div className="flex flex-1 items-center gap-2 group-data-[collapsible=icon]:hidden">
            <div className="grid flex-1 text-left leading-tight">
              <span className="truncate text-sm font-medium">{name}</span>
              <span className="text-muted-foreground truncate text-xs">{subtitle}</span>
            </div>
            <IconDotsVertical className="text-muted-foreground size-4 shrink-0" />
          </div>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
        side={isMobile ? "bottom" : "right"}
        align="end"
        sideOffset={4}
      >
        <DropdownMenuLabel className="p-0 font-normal">
          <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
            <Avatar className="size-8 rounded-md">
              <AvatarImage src={avatarUrl} alt={name} />
              <AvatarFallback className="rounded-md">{getFallback()}</AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium">{name}</span>
              <span className="text-muted-foreground truncate text-xs">{subtitle}</span>
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem asChild>
            <Link to="/account">
              <IconUserCircle />
              个人中心
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/app-setting">
              <IconSettings />
              应用设置
            </Link>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleLogout}>
          <IconLogout />
          退出登录
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
