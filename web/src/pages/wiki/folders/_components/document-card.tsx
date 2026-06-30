import { IconDotsVertical, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { toast } from "sonner";

import type { Document } from "@/api/wiki/document";
import { documentApi } from "@/api/wiki/document";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { getDocumentIconAsset } from "../_shared/file-icons";

export interface DocumentCardProps {
  document: Document;
  onDelete: (document: Document) => void;
  onOpen?: (document: Document) => void;
}

export function DocumentCard({ document, onDelete, onOpen }: DocumentCardProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const iconAsset = getDocumentIconAsset({
    contentType: document.content_type,
    filename: document.filename,
  });

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

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(document);
    setMenuOpen(false);
  };

  return (
    <div
      className="group relative flex flex-col rounded-lg border bg-card p-3 transition-all hover:border-primary hover:shadow-md cursor-pointer"
      onClick={handleClick}
    >
      <div className="flex items-center gap-3 pr-9">
        <div className="rounded-lg bg-muted p-2">
          <img src={iconAsset.src} alt={iconAsset.alt} className="h-6 w-6 object-contain" />
        </div>
        <div className="min-w-0 flex-1">
          <h3
            className="line-clamp-3 text-[15px] leading-5 font-medium text-foreground [overflow-wrap:anywhere]"
            title={document.filename}
          >
            {document.filename}
          </h3>
        </div>
      </div>
      <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="secondary"
            size="icon"
            className="absolute top-2.5 right-2.5 h-7 w-7"
            onClick={handleMenuClick}
          >
            <IconDotsVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleDownload} disabled={downloading}>
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
