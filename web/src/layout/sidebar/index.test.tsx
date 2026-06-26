import {
  Outlet,
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { SidebarProvider } from "@/components/ui/sidebar";
import { useAuthStore } from "@/stores/auth";

import { AppSidebar } from "./index";

async function renderSidebar() {
  const rootRoute = createRootRoute({
    component: () => <Outlet />,
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: () => (
      <SidebarProvider defaultOpen>
        <AppSidebar />
      </SidebarProvider>
    ),
  });

  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });

  const rendered = render(<RouterProvider router={router} />);
  await router.navigate({ to: "/" });
  return rendered;
}

describe("AppSidebar", () => {
  afterEach(() => {
    useAuthStore.getState().clearAuth();
  });

  it("renders the reference-style navigation groups", async () => {
    useAuthStore.setState({
      accessToken: "token",
      refreshToken: "refresh",
      user: {
        id: "1",
        username: "layout-admin",
        email: "layout@example.com",
        full_name: "Layout Admin",
        status: "active",
        roles: ["admin"],
      },
    });

    await renderSidebar();

    expect(await screen.findByText("AI 工作台")).toBeInTheDocument();
    expect(await screen.findByText("知识中心")).toBeInTheDocument();
    expect(await screen.findByText("系统管理")).toBeInTheDocument();
  });

  it("uses the floating sidebar variant", async () => {
    useAuthStore.setState({
      accessToken: "token",
      refreshToken: "refresh",
      user: {
        id: "1",
        username: "layout-admin",
        email: "layout@example.com",
        full_name: "Layout Admin",
        status: "active",
        roles: ["admin"],
      },
    });

    const { container } = await renderSidebar();

    expect(container.querySelector('[data-variant="floating"]')).toBeTruthy();
  });

  it("shows a footer theme switch entry", async () => {
    await renderSidebar();

    expect(await screen.findByText(/浅色|深色/)).toBeInTheDocument();
  });
});
