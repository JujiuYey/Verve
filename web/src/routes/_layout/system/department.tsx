import { createFileRoute } from "@tanstack/react-router";

import { DepartmentPage } from "@/pages/system/department";

export const Route = createFileRoute("/_layout/system/department")({
  component: DepartmentPage,
});
