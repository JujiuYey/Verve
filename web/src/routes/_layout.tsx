import { createFileRoute, redirect } from "@tanstack/react-router";

import { LayoutComponent } from "@/layout/index";
import { useAuthStore } from "@/stores/auth";

export const Route = createFileRoute("/_layout")({
  beforeLoad: () => {
    const { accessToken } = useAuthStore.getState();
    if (!accessToken) {
      throw redirect({ to: "/login" });
    }
  },
  component: LayoutComponent,
});
