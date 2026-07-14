import { createFileRoute } from "@tanstack/react-router";

import { LayoutComponent } from "@/layout/index";

export const Route = createFileRoute("/_layout")({
  component: LayoutComponent,
});
