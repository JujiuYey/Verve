import { createFileRoute, lazyRouteComponent } from "@tanstack/react-router";

import { RoutePendingState } from "../_shared/-route-pending-state";

export const Route = createFileRoute("/_layout/system/model-config")({
  component: lazyRouteComponent(() => import("@/pages/system/model-config"), "ModelConfigPage"),
  pendingComponent: () => <RoutePendingState message="正在加载模型配置..." />,
});
