import { IconUpload } from "@tabler/icons-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

import { documentApi } from "@/api/wiki/document";
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

interface UploadDialogProps {
  open: boolean;
  defaultFolderId?: string;
  folderTree?: FolderTreeNode[];
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
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
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
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
      setSelectedFile(null);
      setSelectedFolderId(defaultFolderId || "");
    }
  }, [open, defaultFolderId]);

  const resolvedFolderTree = folderTree || fallbackFolderTree;
  const folderTreeItems = useMemo(
    () => folderTreeToTreeSelectItems(resolvedFolderTree),
    [resolvedFolderTree],
  );

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const allowedExts = [".md"];
    const ext = file.name.substring(file.name.lastIndexOf(".")).toLowerCase();
    if (!allowedExts.includes(ext)) {
      toast.error("不支持的文件类型", {
        description: "仅支持 .md 文件",
      });
      return;
    }

    setSelectedFile(file);
  };

  const handleUpload = async () => {
    if (!selectedFolderId) {
      toast.error("请选择文件夹");
      return;
    }
    if (!selectedFile) {
      toast.error("请选择文件");
      return;
    }

    try {
      const result = await documentApi.upload(selectedFile, selectedFolderId);
      toast.success("文档上传成功", {
        description: `${result.filename} 已成功上传`,
      });
      onOpenChange(false);
      onSuccess?.();
    } catch {
      // 错误已在拦截器中处理
    } finally {
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
          <DialogDescription>选择文件夹并上传单个 Markdown 文件</DialogDescription>
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
            <label className="text-sm font-medium">文件</label>
            <div
              className="flex items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
              onClick={() => fileInputRef.current?.click()}
            >
              {selectedFile ? (
                <div className="text-center">
                  <p className="text-sm font-medium">{selectedFile.name}</p>
                  <p className="text-xs text-muted-foreground mt-1">点击重新选择</p>
                </div>
              ) : (
                <div className="text-center">
                  <IconUpload className="mx-auto h-8 w-8 text-muted-foreground" />
                  <p className="text-sm text-muted-foreground mt-2">点击选择文件</p>
                  <p className="text-xs text-muted-foreground">支持 .md 文件</p>
                </div>
              )}
            </div>
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              accept=".md"
              onChange={handleFileChange}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={handleUpload} disabled={!selectedFile || !selectedFolderId}>
            <>
              <IconUpload className="mr-2 h-4 w-4" />
              上传
            </>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
