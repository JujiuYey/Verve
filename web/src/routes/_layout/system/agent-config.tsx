import { createFileRoute, lazyRouteComponent } from "@tanstack/react-router";

import { RoutePendingState } from "../_shared/-route-pending-state";

export const Route = createFileRoute("/_layout/system/agent-config")({
  component: lazyRouteComponent(() => import("@/pages/system/agent-config"), "AgentConfigPage"),
  pendingComponent: () => <RoutePendingState message="正在加载 Agent 配置..." />,
});
