import { createFileRoute } from "@tanstack/react-router";

import { RolePage } from "@/pages/system/role";

export const Route = createFileRoute("/_layout/system/role")({
  component: RolePage,
});
