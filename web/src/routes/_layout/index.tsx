import { createFileRoute } from "@tanstack/react-router";

import { FoldersPage } from "@/pages/wiki/folders";

export const Route = createFileRoute("/_layout/")({
  component: FoldersPage,
});
