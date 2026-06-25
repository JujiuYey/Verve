import { createFileRoute } from "@tanstack/react-router";

import { AgentPage } from "@/pages/system/agent";

export const Route = createFileRoute("/_layout/system/agent")({
  component: AgentPage,
});
