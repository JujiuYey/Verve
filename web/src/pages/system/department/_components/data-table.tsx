import {
  IconChevronDown,
  IconChevronRight,
  IconDotsVertical,
  IconEdit,
  IconTrash,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type ExpandedState,
  getExpandedRowModel,
  getFilteredRowModel,
} from "@tanstack/react-table";
import * as React from "react";

import type { Department } from "@/api/system/department";
import { DataTable as GenericDataTable } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface DataTableProps {
  data: Department[];
  loading?: boolean;
  globalFilter?: string;
  onGlobalFilterChange?: (value: string) => void;
  onEdit?: (department: Department) => void;
  onDelete?: (department: Department) => void;
}

export function DataTable({
  data,
  loading = false,
  globalFilter = "",
  onGlobalFilterChange,
  onEdit,
  onDelete,
}: DataTableProps) {
  const [expanded, setExpanded] = React.useState<ExpandedState>(true);

  const columns = React.useMemo<ColumnDef<Department>[]>(
    () => [
      {
        accessorKey: "name",
        header: "部门名称",
        cell: ({ row }) => (
          <div className="flex items-center" style={{ paddingLeft: `${row.depth * 24}px` }}>
            {row.subRows?.length > 0 ? (
              <Button
                variant="ghost"
                size="icon"
                className="size-6 mr-1"
                onClick={row.getToggleExpandedHandler()}
              >
                {row.getIsExpanded() ? (
                  <IconChevronDown className="size-4" />
                ) : (
                  <IconChevronRight className="size-4" />
                )}
              </Button>
            ) : (
              <span className="inline-block w-7" />
            )}
            <span className="font-medium">{row.original.name}</span>
          </div>
        ),
      },
      {
        accessorKey: "description",
        header: "描述",
        cell: ({ row }) => {
          const description = row.original.description || "-";
          return <div className="text-muted-foreground max-w-xs truncate">{description}</div>;
        },
      },
      {
        accessorKey: "created_at",
        header: "创建时间",
        cell: ({ row }) => {
          const date = new Date(row.original.created_at);
          return <div className="text-muted-foreground">{date.toLocaleDateString("zh-CN")}</div>;
        },
      },
      {
        id: "actions",
        header: () => <div className="text-center">操作</div>,
        cell: ({ row }) => (
          <div className="flex items-center justify-center">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className="data-[state=open]:bg-muted text-muted-foreground flex size-8"
                  size="icon"
                >
                  <IconDotsVertical />
                  <span className="sr-only">打开菜单</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-32">
                <DropdownMenuItem onClick={() => onEdit?.(row.original)}>
                  <IconEdit />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem variant="destructive" onClick={() => onDelete?.(row.original)}>
                  <IconTrash />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        ),
      },
    ],
    [onEdit, onDelete],
  );

  return (
    <GenericDataTable<Department>
      columns={columns}
      data={data}
      loading={loading}
      emptyText="暂无数据"
      tableOptions={{
        state: { expanded, globalFilter },
        onExpandedChange: setExpanded,
        onGlobalFilterChange,
        getSubRows: (row) => row.children,
        getRowId: (row) => row.id,
        getExpandedRowModel: getExpandedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
      }}
    />
  );
}
