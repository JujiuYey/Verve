import { createFileRoute } from "@tanstack/react-router";

import { FeynmanWorkbenchPage } from "@/pages/learning/feynman-practice";

export const Route = createFileRoute("/_layout/learn/feynman-practice/$goalId/$objectiveId")({
  component: FeynmanWorkbenchPage,
});
