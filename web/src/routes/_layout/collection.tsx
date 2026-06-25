import { createFileRoute } from "@tanstack/react-router";

import { CollectionPage } from "@/pages/ai/collection";

export const Route = createFileRoute("/_layout/collection")({
  component: CollectionPage,
});
