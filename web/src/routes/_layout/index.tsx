import { createFileRoute } from "@tanstack/react-router";

import { LearningOverviewPage } from "@/pages/learning/overview";

export const Route = createFileRoute("/_layout/")({
  component: LearningOverviewPage,
});
