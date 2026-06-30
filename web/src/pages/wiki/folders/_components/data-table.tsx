import { IconDotsVertical, IconEdit, IconFolder, IconLock, IconTrash } from "@tabler/icons-react";
import { type ColumnDef } from "@tanstack/react-table";
import * as React from "react";

import type { Folder } from "@/api/wiki/folder";
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
  data: Folder[];
  loading?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onEdit?: (folder: Folder) => void;
  onDelete?: (folder: Folder) => void;
  onPermission?: (folder: Folder) => void;
}

export function DataTable({
  data,
  loading = false,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  onEdit,
  onDelete,
  onPermission,
}: DataTableProps) {
  const columns = React.useMemo<ColumnDef<Folder>[]>(
    () => [
      {
        accessorKey: "name",
        header: "文件夹名称",
        cell: ({ row }) => (
          <div className="flex items-center gap-2">
            <IconFolder className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">{row.original.name}</span>
          </div>
        ),
      },
      {
        accessorKey: "description",
        header: "描述",
        cell: ({ row }) => (
          <div className="text-muted-foreground max-w-xs truncate">
            {row.original.description || "-"}
          </div>
        ),
      },
      {
        accessorKey: "created_at",
        header: "创建时间",
        cell: ({ row }) => {
          const date = new Date(row.original.created_at);
          return <div className="text-muted-foreground">{date.toLocaleString("zh-CN")}</div>;
        },
      },
      {
        id: "actions",
        header: () => <div className="text-center">操作</div>,
        cell: ({ row }) => {
          const folder = row.original;
          return (
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
                  <DropdownMenuItem onClick={() => onEdit?.(folder)}>
                    <IconEdit />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onPermission?.(folder)}>
                    <IconLock />
                    权限
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem variant="destructive" onClick={() => onDelete?.(folder)}>
                    <IconTrash />
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          );
        },
      },
    ],
    [onEdit, onDelete, onPermission],
  );

  return (
    <GenericDataTable
      columns={columns}
      data={data}
      loading={loading}
      emptyText="暂无数据"
      pagination={{
        page,
        pageSize,
        total,
        onPageChange,
        onPageSizeChange,
      }}
    />
  );
}
