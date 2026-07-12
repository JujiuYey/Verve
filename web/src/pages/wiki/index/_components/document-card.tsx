import {
  IconAlertCircle,
  IconCircleCheck,
  IconClock,
  IconDatabase,
  IconDotsVertical,
  IconDownload,
  IconLoader2,
  IconTrash,
} from "@tabler/icons-react";
import { useState } from "react";
import { toast } from "sonner";

import { type IndexJobProgress, type IndexJobStatus, ragWikiApi } from "@/api/rag/wiki";
import type { Document } from "@/api/wiki/document";
import { documentApi } from "@/api/wiki/document";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Spinner } from "@/components/ui/spinner";

import { getDocumentIconAsset } from "../_shared/file-icons";

export interface DocumentCardProps {
  document: Document;
  indexJob?: IndexJobProgress;
  onDelete: (document: Document) => void;
  onOpen?: (document: Document) => void;
  onIndexStatusRefresh?: () => void;
  opening?: boolean;
}

const indexStatusMeta: Record<
  IndexJobStatus | "not_started",
  {
    label: string;
    className: string;
    icon: React.ComponentType<{ className?: string }>;
  }
> = {
  not_started: {
    label: "未向量化",
    className: "border-border text-muted-foreground",
    icon: IconClock,
  },
  pending: {
    label: "等待向量化",
    className: "border-amber-200 bg-amber-50 text-amber-700",
    icon: IconClock,
  },
  running: {
    label: "向量化中",
    className: "border-blue-200 bg-blue-50 text-blue-700",
    icon: IconLoader2,
  },
  completed: {
    label: "已向量化",
    className: "border-emerald-200 bg-emerald-50 text-emerald-700",
    icon: IconCircleCheck,
  },
  failed: {
    label: "向量化失败",
    className: "border-destructive/30 bg-destructive/10 text-destructive",
    icon: IconAlertCircle,
  },
  superseded: {
    label: "版本已过期",
    className: "border-border text-muted-foreground",
    icon: IconClock,
  },
};

export function DocumentCard({
  document,
  indexJob,
  onDelete,
  onOpen,
  onIndexStatusRefresh,
  opening,
}: DocumentCardProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [indexing, setIndexing] = useState(false);
  const iconAsset = getDocumentIconAsset({
    contentType: document.content_type,
    filename: document.filename,
  });
  const indexStatus = indexing ? "running" : (indexJob?.status ?? "not_started");
  const statusMeta = indexStatusMeta[indexStatus];
  const StatusIcon = statusMeta.icon;
  const statusLabel =
    indexStatus === "completed" && indexJob?.chunk_count
      ? `${statusMeta.label} · ${indexJob.chunk_count} 片段`
      : statusMeta.label;
  const indexActionDisabled = indexing || indexStatus === "pending" || indexStatus === "running";
  const indexActionLabel =
    indexStatus === "pending"
      ? "排队中"
      : indexStatus === "running"
        ? "解析中..."
        : indexStatus === "not_started"
          ? "解析"
          : "重新解析";

  const handleClick = () => {
    onOpen?.(document);
  };

  const handleMenuClick = (e: React.MouseEvent) => {
    e.stopPropagation();
  };

  const handleDownload = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setDownloading(true);
    try {
      const res = await documentApi.download(document.id);
      // 打开下载链接
      window.open(res.download_url, "_blank");
    } catch (error) {
      toast.error("下载失败");
      console.error("下载失败:", error);
    } finally {
      setDownloading(false);
      setMenuOpen(false);
    }
  };

  const handleIndexDocument = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (indexActionDisabled) return;

    setIndexing(true);
    setMenuOpen(false);
    try {
      await ragWikiApi.indexDocument(document.id);
      toast.success("解析完成");
    } catch (error) {
      console.error("解析失败:", error);
    } finally {
      setIndexing(false);
      onIndexStatusRefresh?.();
    }
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(document);
    setMenuOpen(false);
  };

  return (
    <div
      className="group relative cursor-pointer rounded-lg border bg-card p-4 transition-all hover:border-primary hover:shadow-md"
      onClick={handleClick}
    >
      <div className="flex items-center gap-3 pr-8">
        <div className="rounded-md bg-muted p-2">
          {opening ? (
            <Spinner className="size-6" />
          ) : (
            <img src={iconAsset.src} alt={iconAsset.alt} className="size-6 object-contain" />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <h3
            className="line-clamp-3 text-[15px] leading-5 font-medium text-foreground [overflow-wrap:anywhere]"
            title={document.filename}
          >
            {document.filename}
          </h3>
          <Badge
            variant="outline"
            className={`mt-2 h-5 gap-1 px-1.5 font-normal ${statusMeta.className}`}
            title={indexJob?.error_message || statusLabel}
          >
            <StatusIcon className={indexStatus === "running" ? "animate-spin" : ""} />
            {statusLabel}
          </Badge>
        </div>
      </div>
      <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="secondary"
            size="icon"
            className="absolute top-1/2 right-2 h-8 w-8 -translate-y-1/2"
            onClick={handleMenuClick}
          >
            <IconDotsVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleIndexDocument} disabled={indexActionDisabled}>
            <IconDatabase className="mr-2 h-4 w-4" />
            {indexActionLabel}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={handleDownload} disabled={downloading}>
            <IconDownload className="mr-2 h-4 w-4" />
            {downloading ? "下载中..." : "下载"}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleDelete} className="text-destructive">
            <IconTrash className="mr-2 h-4 w-4" />
            删除
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
