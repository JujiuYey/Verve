import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_layout/learn/feynman-practice/$documentId/")({
  beforeLoad: ({ params, search }) => {
    throw redirect({
      to: "/learn/feynman-practice/$documentId/listener",
      params: { documentId: params.documentId },
      search: { sessionId: search.sessionId },
      replace: true,
    });
  },
});
