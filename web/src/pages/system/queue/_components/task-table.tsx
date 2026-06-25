import { type ColumnDef } from "@tanstack/react-table";
import * as React from "react";

import type { TaskInfo } from "@/api/system/queue";
import { DataTable } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";

interface TaskTableProps {
  tasks: TaskInfo[];
  loading?: boolean;
  state: string;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

function getStateBadgeVariant(state: string) {
  const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
    pending: "secondary",
    active: "default",
    scheduled: "outline",
    retry: "destructive",
    archived: "secondary",
    completed: "default",
  };
  return variants[state] || "default";
}

export function TaskTable({
  tasks,
  loading = false,
  state,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
}: TaskTableProps) {
  const columns = React.useMemo<ColumnDef<TaskInfo>[]>(() => {
    const cols: ColumnDef<TaskInfo>[] = [
      {
        accessorKey: "id",
        header: "任务ID",
        cell: ({ row }) => (
          <span className="font-mono text-xs">
            {row.original.id.substring(0, 8)}
            ...
          </span>
        ),
      },
      {
        accessorKey: "type",
        header: "类型",
      },
      {
        accessorKey: "queue",
        header: "队列",
      },
    ];

    if (state !== "scheduled") {
      cols.push({
        id: "retry_info",
        header: "重试次数",
        cell: ({ row }) => (
          <span>
            {row.original.retried} /{row.original.max_retry}
          </span>
        ),
      });
    }

    if (state === "pending" || state === "active") {
      cols.push({
        accessorKey: "state",
        header: "状态",
        cell: ({ row }) => (
          <Badge variant={getStateBadgeVariant(row.original.state)}>{row.original.state}</Badge>
        ),
      });
    }

    if (state === "retry" || state === "archived") {
      cols.push({
        accessorKey: "last_error",
        header: "错误信息",
        cell: ({ row }) => (
          <div className="max-w-xs truncate text-xs text-red-500">
            {row.original.last_error || "-"}
          </div>
        ),
      });
    }

    if (state === "pending" || state === "scheduled" || state === "retry") {
      cols.push({
        accessorKey: "next_process",
        header: "下次执行",
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">{row.original.next_process || "-"}</span>
        ),
      });
    }

    return cols;
  }, [state]);

  return (
    <DataTable
      columns={columns}
      data={tasks}
      loading={loading}
      emptyText="暂无任务"
      pagination={
        total > 0
          ? {
              page,
              pageSize,
              total,
              pageSizeOptions: [10, 20, 30, 50],
              onPageChange,
              onPageSizeChange,
            }
          : undefined
      }
    />
  );
}
