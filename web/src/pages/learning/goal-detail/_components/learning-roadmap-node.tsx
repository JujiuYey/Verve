import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import {
  ChevronDownIcon,
  ChevronRightIcon,
  Clock3Icon,
  FileTextIcon,
  FolderOpenIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type RoadmapNodeData = {
  label: string;
  duration: string;
  difficulty: string;
  status: "planned" | "active" | "completed";
  kind: "folder" | "objective";
  stageId: string;
  folderPath?: string;
  isCollapsed?: boolean;
  childCount?: number;
  objectiveId?: string;
};

const statusMeta: Record<RoadmapNodeData["status"], { label: string; dotClassName: string }> = {
  planned: {
    label: "待开始",
    dotClassName: "bg-muted-foreground/50",
  },
  active: {
    label: "进行中",
    dotClassName: "bg-primary",
  },
  completed: {
    label: "已完成",
    dotClassName: "bg-emerald-500",
  },
};

export function LearningRoadmapNode({ data, selected }: NodeProps<Node<RoadmapNodeData>>) {
  const meta = statusMeta[data.status];
  const isObjective = data.kind === "objective";
  const KindIcon = isObjective ? FileTextIcon : FolderOpenIcon;
  const ToggleIcon = data.isCollapsed ? ChevronRightIcon : ChevronDownIcon;

  return (
    <div
      className={cn(
        "w-56 rounded-2xl border bg-card p-4 shadow-sm transition-colors",
        isObjective && "rounded-xl border-border bg-background",
        selected && "border-primary ring-2 ring-primary/15",
      )}
    >
      <Handle
        type="target"
        position={Position.Left}
        className="!size-3 !border-2 !border-background !bg-primary"
      />
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className={cn("size-2 rounded-full", meta.dotClassName)} />
            <KindIcon className="size-3.5" />
            {meta.label}
            {!isObjective ? (
              <span className="ml-auto inline-flex items-center gap-1">
                <ToggleIcon className="size-3.5" />
                {data.childCount ? `${data.childCount} 项` : null}
              </span>
            ) : null}
          </div>
          <div className="text-sm font-semibold leading-6">{data.label}</div>
        </div>
        <Badge variant="outline">{data.difficulty}</Badge>
      </div>
      <div className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
        <Clock3Icon className="size-3.5" />
        {data.duration}
      </div>
      <Handle
        type="source"
        position={Position.Right}
        className="!size-3 !border-2 !border-background !bg-primary"
      />
    </div>
  );
}
