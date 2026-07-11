import { createFileRoute } from "@tanstack/react-router";

import { FeynmanWorkbenchPage } from "@/pages/learning/feynman-practice";

interface SearchSchema {
  sessionId?: string;
}

export const Route = createFileRoute("/_layout/learn/feynman-practice/$documentId")({
  component: FeynmanWorkbenchPage,
  validateSearch: (search: Record<string, unknown>): SearchSchema => ({
    sessionId: typeof search.sessionId === "string" ? search.sessionId : undefined,
  }),
});
