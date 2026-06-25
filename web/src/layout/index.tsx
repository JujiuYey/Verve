import { Outlet } from "@tanstack/react-router";
import { useEffect } from "react";

import { authApi } from "@/api/auth/auth";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useAuthStore } from "@/stores/auth";

import { AppSidebar } from "./sidebar";

export function LayoutComponent() {
  const { user, setUser, accessToken } = useAuthStore();

  useEffect(() => {
    // 如果有 token 但没有用户信息，则获取用户信息
    if (accessToken && !user) {
      authApi
        .getCurrentUser()
        .then((userInfo) => {
          setUser(userInfo);
        })
        .catch(() => {
          // 获取用户信息失败，可能是 token 过期
          console.error("获取用户信息失败");
        });
    }
  }, [accessToken, user, setUser]);

  return (
    <SidebarProvider defaultOpen>
      <AppSidebar />
      <SidebarInset>
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  );
}
