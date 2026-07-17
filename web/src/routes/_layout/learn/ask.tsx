import { createFileRoute } from "@tanstack/react-router";

import { KnowledgeQAPage } from "@/pages/learning/ask";

export const Route = createFileRoute("/_layout/learn/ask")({
  component: KnowledgeQAPage,
});
