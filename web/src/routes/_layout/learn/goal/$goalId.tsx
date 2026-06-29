import { createFileRoute } from "@tanstack/react-router";

import { GoalDetailPage } from "@/pages/learning/goal-detail";

export const Route = createFileRoute("/_layout/learn/goal/$goalId")({
  component: GoalDetailPage,
});
