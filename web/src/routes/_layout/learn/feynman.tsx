import { createFileRoute } from "@tanstack/react-router";

import { FeynmanExercisePage } from "@/pages/learning/feynman";

export const Route = createFileRoute("/_layout/learn/feynman")({
  component: FeynmanExercisePage,
});
