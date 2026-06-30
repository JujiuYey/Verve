import {
  IconDownload,
  IconEdit,
  IconFileText,
  IconTrash,
} from "@tabler/icons-react";
import type { ColumnDef } from "@tanstack/react-table";
import * as React from "react";

import type { Document } from "@/api/wiki/document";
import { DataTable } from "@/components/sag-ui/data-table";
import { Button } from "@/components/ui/button";

function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(2)} ${sizes[i]}`;
}

interface DataTableProps {
  data: Document[];
  loading?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onDownload?: (doc: Document) => void;
  onDelete?: (doc: Document) => void;
  onRowClick?: (doc: Document) => void;
  onEditCanvas?: (doc: Document) => void;
}

export function DocumentDataTable({
  data,
  loading = false,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  onDownload,
  onDelete,
  onRowClick,
  onEditCanvas,
}: DataTableProps) {
  const columns = React.useMemo<ColumnDef<Document>[]>(
    () => [
      {
        accessorKey: "filename",
        header: () => <div className="text-center">文件名</div>,
        cell: ({ row }) => (
          <div className="flex items-center gap-2">
            <IconFileText className="h-4 w-4 text-muted-foreground shrink-0" />
            <span className="max-w-75 truncate font-medium" title={row.original.filename}>
              {row.original.filename}
            </span>
          </div>
        ),
      },
      {
        accessorKey: "file_size",
        header: () => <div className="text-center">大小</div>,
        cell: ({ row }) => <div className="w-20">{formatFileSize(row.original.file_size)}</div>,
      },
      {
        accessorKey: "created_at",
        header: () => <div className="text-center">上传时间</div>,
        cell: ({ row }) => {
          const date = new Date(row.original.created_at);
          return (
            <div className="text-muted-foreground w-37.5">{date.toLocaleDateString("zh-CN")}</div>
          );
        },
      },
      {
        id: "actions",
        header: () => <div className="text-center">操作</div>,
        cell: ({ row }) => {
          const doc = row.original;

          return (
            <div className="flex items-center justify-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={(e) => {
                  e.stopPropagation();
                  onEditCanvas?.(doc);
                }}
              >
                <IconEdit className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={(e) => {
                  e.stopPropagation();
                  onDownload?.(doc);
                }}
              >
                <IconDownload className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete?.(doc);
                }}
              >
                <IconTrash className="h-4 w-4 text-destructive" />
              </Button>
            </div>
          );
        },
      },
    ],
    [onDownload, onDelete, onEditCanvas],
  );

  return (
    <DataTable
      columns={columns}
      data={data}
      loading={loading}
      pagination={{
        page,
        pageSize,
        total,
        pageSizeOptions: [10, 20, 30, 50],
        onPageChange,
        onPageSizeChange,
      }}
      onRowClick={onRowClick}
    />
  );
}
