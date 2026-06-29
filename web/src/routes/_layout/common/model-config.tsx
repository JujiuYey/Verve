import { createFileRoute } from "@tanstack/react-router";

import { ModelConfigPage } from "@/pages/ai/model-config";

export const Route = createFileRoute("/_layout/common/model-config")({
  component: ModelConfigPage,
});
