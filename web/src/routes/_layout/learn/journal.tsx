import { createFileRoute } from "@tanstack/react-router";

import { JournalPage } from "@/pages/learning/journal";

export const Route = createFileRoute("/_layout/learn/journal")({
  component: JournalPage,
});
