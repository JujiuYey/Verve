import { createFileRoute, Outlet } from "@tanstack/react-router";

interface SearchSchema {
  sessionId?: string;
}

export const Route = createFileRoute("/_layout/learn/feynman-practice/$documentId")({
  component: Outlet,
  validateSearch: (search: Record<string, unknown>): SearchSchema => ({
    sessionId: typeof search.sessionId === "string" ? search.sessionId : undefined,
  }),
});
