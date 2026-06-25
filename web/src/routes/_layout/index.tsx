import { createFileRoute } from "@tanstack/react-router";

import { DashboardPage } from "@/pages/dashboard";

export const Route = createFileRoute("/_layout/")({
  component: DashboardPage,
});
