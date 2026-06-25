import { render, screen } from "@testing-library/react";
import {
  Outlet,
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { describe, expect, it } from "vitest";

import { LoginPage } from "./index";

async function renderLoginPage() {
  const rootRoute = createRootRoute({
    component: () => <Outlet />,
  });

  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/login",
    component: LoginPage,
  });

  const router = createRouter({
    routeTree: rootRoute.addChildren([loginRoute]),
    history: createMemoryHistory({ initialEntries: ["/login"] }),
  });

  render(<RouterProvider router={router} />);
  await router.navigate({ to: "/login" });
}

describe("LoginPage", () => {
  it("renders the reference-style marketing panel content", async () => {
    await renderLoginPage();

    expect((await screen.findAllByText("沉淀结构化知识资产")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("SAG Wiki")).length).toBeGreaterThan(0);
  });
});
