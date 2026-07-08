import { createFileRoute } from "@tanstack/react-router";

import { WikiIndexPage } from "@/pages/wiki/index";

export const Route = createFileRoute("/_layout/wiki/")({
  component: WikiIndexPage,
});
