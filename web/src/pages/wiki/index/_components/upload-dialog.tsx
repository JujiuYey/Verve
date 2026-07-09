import { IconFile, IconUpload, IconX } from "@tabler/icons-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

import { documentApi, type Document } from "@/api/wiki/document";
import { folderApi, type FolderTreeNode } from "@/api/wiki/folder";
import { TreeSelect, type TreeSelectItem } from "@/components/sag-ui/tree-select";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";

interface UploadDialogProps {
  open: boolean;
  defaultFolderId?: string;
  folderTree?: FolderTreeNode[];
  onOpenChange: (open: boolean) => void;
  onSuccess?: (document: Document, folderId: string) => void;
}

const MAX_CONCURRENT_UPLOADS = 4;
const ALLOWED_EXTENSIONS = [".md"];

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
}

function folderTreeToTreeSelectItems(nodes: FolderTreeNode[]): TreeSelectItem<FolderTreeNode>[] {
  return nodes.map((node) => ({
    value: node.id,
    label: node.name,
    node,
    children: node.children.length > 0 ? folderTreeToTreeSelectItems(node.children) : undefined,
  }));
}

export function UploadDialog({
  open,
  defaultFolderId,
  folderTree,
  onOpenChange,
  onSuccess,
}: UploadDialogProps) {
  const [fallbackFolderTree, setFallbackFolderTree] = useState<FolderTreeNode[]>([]);
  const [selectedFolderId, setSelectedFolderId] = useState<string>("");
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);
  const [uploadedCount, setUploadedCount] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open || folderTree) return;
    const load = async () => {
      try {
        const tree = await folderApi.tree();
        setFallbackFolderTree(tree || []);
      } catch {
        toast.error("加载文件夹列表失败");
      }
    };
    void load();
  }, [open, folderTree]);

  useEffect(() => {
    if (open && defaultFolderId) {
      setSelectedFolderId(defaultFolderId);
    }
  }, [open, defaultFolderId]);

  useEffect(() => {
    if (!open) {
      setSelectedFiles([]);
      setUploadedCount(0);
      setSelectedFolderId(defaultFolderId || "");
    }
  }, [open, defaultFolderId]);

  const resolvedFolderTree = folderTree || fallbackFolderTree;
  const folderTreeItems = useMemo(
    () => folderTreeToTreeSelectItems(resolvedFolderTree),
    [resolvedFolderTree],
  );

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;

    const supportedFiles = files.filter((file) => {
      const ext = file.name.substring(file.name.lastIndexOf(".")).toLowerCase();
      return ALLOWED_EXTENSIONS.includes(ext);
    });
    const rejectedCount = files.length - supportedFiles.length;

    if (rejectedCount > 0) {
      toast.error("不支持的文件类型", {
        description: `已忽略 ${rejectedCount} 个非 .md 文件`,
      });
    }

    setSelectedFiles((prev) => [...prev, ...supportedFiles]);
  };

  const handleRemoveFile = (index: number) => {
    setSelectedFiles((files) => files.filter((_, i) => i !== index));
  };

  const handleUpload = async () => {
    if (!selectedFolderId) {
      toast.error("请选择文件夹");
      return;
    }
    if (selectedFiles.length === 0) {
      toast.error("请选择文件");
      return;
    }

    const total = selectedFiles.length;
    setUploading(true);
    setUploadedCount(0);
    let successCount = 0;
    let failedCount = 0;

    const results: Array<Document | null> = Array.from(
      { length: total },
      () => null,
    );
    const completed: boolean[] = Array.from({ length: total }, () => false);
    let nextEmitIndex = 0;

    const tryEmit = () => {
      while (nextEmitIndex < total && completed[nextEmitIndex]) {
        const result = results[nextEmitIndex];
        if (result) {
          successCount += 1;
          onSuccess?.(result, selectedFolderId);
        } else {
          failedCount += 1;
        }
        nextEmitIndex += 1;
      }
    };

    try {
      const uploadOne = async (index: number) => {
        try {
          const result = await documentApi.upload(
            selectedFiles[index],
            selectedFolderId,
          );
          results[index] = result;
        } catch {
          results[index] = null;
        }
        completed[index] = true;
        setUploadedCount((count) => count + 1);
        tryEmit();
      };

      const workerCount = Math.min(MAX_CONCURRENT_UPLOADS, total);
      await Promise.all(
        Array.from({ length: workerCount }, async (_, workerIdx) => {
          for (let i = workerIdx; i < total; i += workerCount) {
            await uploadOne(i);
          }
        }),
      );

      if (successCount > 0) {
        toast.success("文档上传完成", {
          description:
            failedCount > 0
              ? `成功 ${successCount} 个，失败 ${failedCount} 个`
              : `成功上传 ${successCount} 个文件`,
        });
      }
      if (successCount === 0 && failedCount > 0) {
        toast.error("文档上传失败", {
          description: `${failedCount} 个文件上传失败`,
        });
      }
      onOpenChange(false);
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>上传文档</DialogTitle>
          <DialogDescription>选择文件夹并上传 Markdown 文件</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">文件夹</label>
            <TreeSelect
              items={folderTreeItems}
              value={selectedFolderId}
              onValueChange={(value) => setSelectedFolderId(value)}
              placeholder="请选择文件夹"
              className="w-full"
              emptyMessage="暂无文件夹"
            />
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium">文件</label>
              {selectedFiles.length > 0 && (
                <Button
                  variant="ghost"
                  size="xs"
                  disabled={uploading}
                  onClick={() => fileInputRef.current?.click()}
                >
                  <IconUpload className="h-3 w-3" />
                  添加文件
                </Button>
              )}
            </div>
            {selectedFiles.length === 0 ? (
              <div
                className="flex items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                onClick={() => fileInputRef.current?.click()}
              >
                <div className="text-center">
                  <IconUpload className="mx-auto h-8 w-8 text-muted-foreground" />
                  <p className="text-sm text-muted-foreground mt-2">点击选择文件</p>
                  <p className="text-xs text-muted-foreground">支持 .md 文件</p>
                </div>
              </div>
            ) : (
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground">
                  已选择 {selectedFiles.length} 个文件
                </p>
                <ScrollArea className="h-48 rounded-md border">
                  <div className="p-1 space-y-0.5">
                    {selectedFiles.map((file, index) => (
                      <div
                        key={`${file.name}-${index}`}
                        className="group flex items-center gap-2 rounded px-2 py-1.5 hover:bg-muted/60"
                      >
                        <IconFile className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span
                          className="flex-1 truncate text-sm"
                          title={file.name}
                        >
                          {file.name}
                        </span>
                        <span className="shrink-0 text-xs text-muted-foreground tabular-nums">
                          {formatFileSize(file.size)}
                        </span>
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          disabled={uploading}
                          onClick={() => handleRemoveFile(index)}
                          aria-label={`移除 ${file.name}`}
                        >
                          <IconX className="h-3 w-3" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </div>
            )}
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              accept=".md"
              multiple
              onChange={handleFileChange}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={uploading}>
            取消
          </Button>
          <Button
            onClick={handleUpload}
            disabled={uploading || selectedFiles.length === 0 || !selectedFolderId}
          >
            <>
              <IconUpload className="mr-2 h-4 w-4" />
              {uploading ? `上传中 ${uploadedCount}/${selectedFiles.length}` : "上传"}
            </>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
