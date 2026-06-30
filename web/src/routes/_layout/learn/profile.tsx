import { createFileRoute } from "@tanstack/react-router";

import { ProfilePage } from "@/pages/learning/profile";

export const Route = createFileRoute("/_layout/learn/profile")({
  component: ProfilePage,
});
