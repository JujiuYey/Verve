import { IconDotsVertical, IconEdit, IconKey, IconTrash } from "@tabler/icons-react";
import { type ColumnDef } from "@tanstack/react-table";
import * as React from "react";

import type { User } from "@/api/system/user";
import { DataTable as SagDataTable } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface DataTableProps {
  data: User[];
  loading?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onEdit?: (user: User) => void;
  onDelete?: (user: User) => void;
  onResetPassword?: (user: User) => void;
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
  onResetPassword,
}: DataTableProps) {
  const columns = React.useMemo<ColumnDef<User>[]>(
    () => [
      {
        accessorKey: "username",
        header: "用户名",
        cell: ({ row }) => <span className="font-medium">{row.original.username}</span>,
      },
      {
        accessorKey: "full_name",
        header: "姓名",
        cell: ({ row }) => <span>{row.original.full_name || "-"}</span>,
      },
      {
        accessorKey: "email",
        header: "邮箱",
        cell: ({ row }) => <span className="text-muted-foreground">{row.original.email}</span>,
      },
      {
        accessorKey: "roles",
        header: "角色",
        cell: ({ row }) => {
          const roles = row.original.roles;
          if (!roles || roles.length === 0) return <span className="text-muted-foreground">-</span>;
          return (
            <div className="flex gap-1 flex-wrap">
              {roles.map((role) => (
                <Badge key={role.id} variant="outline" className="text-xs">
                  {role.name}
                </Badge>
              ))}
            </div>
          );
        },
      },
      {
        accessorKey: "status",
        header: "状态",
        cell: ({ row }) => {
          const isActive = row.original.status === "active";
          return (
            <Badge variant={isActive ? "default" : "secondary"}>{isActive ? "启用" : "禁用"}</Badge>
          );
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
                <DropdownMenuItem onClick={() => onResetPassword?.(row.original)}>
                  <IconKey />
                  重置密码
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
    [onEdit, onDelete, onResetPassword],
  );

  return (
    <SagDataTable<User>
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
