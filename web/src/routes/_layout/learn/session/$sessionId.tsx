import { createFileRoute } from "@tanstack/react-router";

import { SessionPage } from "@/pages/learning/session";

export const Route = createFileRoute("/_layout/learn/session/$sessionId")({
  component: SessionPage,
});
