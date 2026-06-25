import {
  IconDownload,
  IconEdit,
  IconFileSearch,
  IconFileText,
  IconLoader2,
  IconTrash,
} from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import * as React from "react";

import type { Document } from "@/api/wiki/document";
import { documentApi } from "@/api/wiki/document";
import { DataTable } from "@/components/sag-ui/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(2)} ${sizes[i]}`;
}

function getStatusText(status: Document["status"]) {
  switch (status) {
    case "pending":
      return "待处理";
    case "processing":
      return "处理中";
    case "completed":
      return "已完成";
    case "failed":
      return "失败";
    default:
      return "未知";
  }
}

function getStatusClass(status: Document["status"]) {
  switch (status) {
    case "pending":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
    case "processing":
      return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300";
    case "completed":
      return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
    case "failed":
      return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300";
  }
}

// ChunkDetail component for expanded row
const ChunkDetail = ({ docId }: { docId: string }) => {
  const { data: chunksData, isLoading } = useQuery({
    queryKey: ["chunks", docId],
    queryFn: () => documentApi.getChunks(docId),
    enabled: !!docId,
  });
  console.log("🚀 ~ ChunkDetail ~ chunksData:", chunksData);

  if (isLoading) return <div className="p-2 text-sm text-muted-foreground">加载中...</div>;
  if (!chunksData?.chunks?.length)
    return <div className="p-2 text-sm text-muted-foreground">暂无 chunk 数据</div>;

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 p-2 max-h-[60vh] overflow-y-auto">
      {chunksData.chunks.map((chunk) => (
        <div key={chunk.ChunkId} className="border rounded p-2 text-sm">
          <div className="flex justify-between items-start">
            <span className="font-medium">
              Chunk
              {chunk.ChunkIndex + 1}
            </span>
            <span className="text-muted-foreground text-xs">
              {chunk.ChunkSize} 字符 |{chunk.VectorDim} 维
            </span>
          </div>
          <p className="mt-1 text-muted-foreground line-clamp-3">{chunk.ChunkText}</p>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="link" size="sm" className="mt-1 h-auto p-0">
                查看完整内容
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>
                  Chunk
                  {chunk.ChunkIndex + 1}
                </DialogTitle>
              </DialogHeader>
              <div className="mt-4">
                <p className="whitespace-pre-wrap text-sm">{chunk.ChunkText}</p>
                <div className="mt-4 text-xs text-muted-foreground">
                  <p>
                    字符数:
                    {chunk.ChunkSize}
                  </p>
                  <p>
                    向量维度:
                    {chunk.VectorDim}
                  </p>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      ))}
    </div>
  );
};

interface DataTableProps {
  data: Document[];
  loading?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onProcess?: (doc: Document) => void;
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
  onProcess,
  onDownload,
  onDelete,
  onRowClick,
  onEditCanvas,
}: DataTableProps) {
  const [selectedDocForChunks, setSelectedDocForChunks] = React.useState<Document | null>(null);

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
        accessorKey: "status",
        header: () => <div className="text-center">状态</div>,
        cell: ({ row }) => {
          const status = row.original.status;
          return (
            <Badge variant="secondary" className={getStatusClass(status)}>
              {getStatusText(status)}
            </Badge>
          );
        },
      },
      {
        accessorKey: "chunk_count",
        header: () => <div className="text-center">分块数</div>,
        cell: ({ row }) => <div className="text-muted-foreground">{row.original.chunk_count}</div>,
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
          const showProcessBtn = doc.status === "pending" || doc.status === "failed";
          const showChunksBtn = doc.status === "completed";

          return (
            <div className="flex items-center justify-center gap-1">
              {showProcessBtn && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={(e) => {
                    e.stopPropagation();
                    onProcess?.(doc);
                  }}
                >
                  <IconLoader2 className="h-4 w-4 text-amber-500" />
                </Button>
              )}
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
              {showChunksBtn && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedDocForChunks(doc);
                  }}
                >
                  <IconFileSearch className="h-4 w-4" />
                </Button>
              )}
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
    [onProcess, onDownload, onDelete, onEditCanvas],
  );

  return (
    <>
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
        onRowClick={
          onRowClick
            ? (doc) => {
                if (doc.status === "completed") onRowClick(doc);
              }
            : undefined
        }
      />

      {/* Chunk Detail Dialog */}
      <Dialog open={!!selectedDocForChunks} onOpenChange={() => setSelectedDocForChunks(null)}>
        <DialogContent className="max-w-[70vw]">
          <DialogHeader>
            <DialogTitle>
              {selectedDocForChunks?.filename}
              {" - "}
              Chunk 详情
            </DialogTitle>
          </DialogHeader>
          {selectedDocForChunks && <ChunkDetail docId={selectedDocForChunks.id} />}
        </DialogContent>
      </Dialog>
    </>
  );
}
