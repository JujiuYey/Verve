import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_layout/learn/feynman")({
  beforeLoad: () => {
    throw redirect({ to: "/learn/ask", replace: true });
  },
});
