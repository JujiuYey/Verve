import { Badge } from "@/components/ui/badge";

import type { AgentStatus } from "../agent-definitions";

export function StatusBadge({ status }: { status: AgentStatus }) {
  if (status === "ready") return <Badge variant="default">就绪</Badge>;
  if (status === "partial") return <Badge variant="secondary">部分</Badge>;
  return <Badge variant="outline">未配置</Badge>;
}
