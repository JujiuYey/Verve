import { createFileRoute } from "@tanstack/react-router";

import { QueuePage } from "@/pages/system/queue";

export const Route = createFileRoute("/_layout/system/queue")({
  component: QueuePage,
});
