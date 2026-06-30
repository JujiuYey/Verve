import { IconRefresh, IconSearch, IconUpload } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { type Document, documentApi } from "@/api/wiki/document";
import { folderApi, type FolderTreeNode } from "@/api/wiki/folder";
import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { FolderTree } from "@/pages/wiki/folders/_components/folder-tree";

import { DocumentDataTable } from "./_components/data-table";
import { UploadDialog } from "./_components/upload-dialog";

export function DocumentsPage() {
  const navigate = useNavigate();

  // 知识库筛选
  const [activeKbId, setActiveKbId] = useState<string>();

  // 文件夹树相关状态
  const [folderTreeData, setFolderTreeData] = useState<FolderTreeNode[]>([]);

  // 文档数据
  const [data, setData] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState("");

  // 上传弹窗
  const [uploadOpen, setUploadOpen] = useState(false);
  const [downloadingIds, setDownloadingIds] = useState<Set<string>>(new Set());

  // 删除确认弹窗状态
  const [deleteTarget, setDeleteTarget] = useState<Document | null>(null);

  // 加载文档列表
  const loadDocuments = useCallback(async () => {
    if (!activeKbId) return;

    setLoading(true);
    try {
      const res = await documentApi.page({
        page_size: pageSize,
        page,
        name: searchKeyword || undefined,
        folder_id: activeKbId,
      });
      setData(res.data || []);
      setTotal(res.total || 0);
    } catch {
      toast.error("加载文档列表失败");
    } finally {
      setLoading(false);
    }
  }, [activeKbId, page, pageSize, searchKeyword]);

  // 知识库或分页变化时重新加载
  useEffect(() => {
    loadDocuments();
  }, [loadDocuments]);

  // 加载文件夹树
  const loadFolderTree = useCallback(async () => {
    try {
      const res = await folderApi.tree();
      setFolderTreeData(res || []);
    } catch {
      toast.error("加载文件夹树失败");
    }
  }, []);

  // 初始化加载文件夹树
  useEffect(() => {
    void loadFolderTree();
  }, [loadFolderTree]);

  // 树形导航选择文件夹
  const handleTreeSelect = useCallback((folder: FolderTreeNode | null) => {
    if (folder === null) {
      setActiveKbId(undefined);
      return;
    }
    setActiveKbId(folder.id);
    setPage(1);
    setSearchKeyword("");
  }, []);

  // 切换知识库时重置分页
  // const handleKbChange = (id: string) => {
  //   setActiveKbId(id);
  //   setPage(1);
  //   setSearchKeyword("");
  // };

  // 搜索
  const handleSearch = () => {
    setPage(1);
    loadDocuments();
  };

  // 页大小变化
  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPage(1);
  };

  // 下载文档
  const handleDownload = async (doc: Document) => {
    if (downloadingIds.has(doc.id)) return;

    setDownloadingIds((prev) => new Set(prev).add(doc.id));
    try {
      const res = await documentApi.download(doc.id);
      window.open(res.download_url, "_blank");
      toast.success("下载已开始", {
        description: `${doc.filename} 下载链接已生成`,
      });
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setDownloadingIds((prev) => {
        const next = new Set(prev);
        next.delete(doc.id);
        return next;
      });
    }
  };

  // 删除文档
  const handleDelete = (doc: Document) => {
    setDeleteTarget(doc);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await documentApi.delete(deleteTarget.id);
      toast.success("删除成功", {
        description: `${deleteTarget.filename} 已被删除`,
      });
      loadDocuments();
    } catch {
      // 错误已在拦截器中处理
    }
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <ResizablePanelGroup orientation="horizontal" className="min-h-0 flex-1 overflow-hidden">
        {/* 左侧文件夹树形导航 */}
        <ResizablePanel
          id="folder-tree-panel"
          defaultSize="20%"
          minSize="15%"
          maxSize="30%"
          className="min-w-0 overflow-y-auto"
        >
          <FolderTree data={folderTreeData} selectedId={activeKbId} onSelect={handleTreeSelect} />
        </ResizablePanel>

        <ResizableHandle withHandle />

        {/* 右侧主内容区 */}
        <ResizablePanel
          id="folder-content-panel"
          defaultSize="80%"
          minSize="50%"
          className="min-w-0 overflow-y-auto"
        >
          {!activeKbId ? (
            <div className="h-full flex items-center justify-center">
              <div className="text-center space-y-4 max-w-md px-6">
                <div className="flex justify-center">
                  <div className="rounded-full bg-muted p-6">
                    <IconSearch className="h-12 w-12 text-muted-foreground" />
                  </div>
                </div>
                <div className="space-y-2">
                  <h2 className="text-2xl font-semibold">请选择知识库</h2>
                  <p className="text-muted-foreground">请从左侧选择一个知识库来查看和管理文档</p>
                </div>
              </div>
            </div>
          ) : (
            <>
              {/* 头部 */}
              <div className="border-b">
                <div className="flex items-center justify-between p-6">
                  <div>
                    <h1 className="text-2xl font-bold">文档管理</h1>
                    <p className="text-sm text-muted-foreground mt-1">上传和管理您的知识库文档</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button size="sm" variant="outline" onClick={loadDocuments}>
                      <IconRefresh className="h-4 w-4" />
                    </Button>
                    <Button onClick={() => setUploadOpen(true)}>
                      <IconUpload className="mr-2 h-4 w-4" />
                      上传文档
                    </Button>
                  </div>
                </div>

                {/* 搜索栏 */}
                <div className="px-6 pb-4">
                  <div className="flex gap-2">
                    <div className="relative flex-1 max-w-sm">
                      <IconSearch className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="搜索文档名称..."
                        value={searchKeyword}
                        onChange={(e) => setSearchKeyword(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                        className="pl-9"
                      />
                    </div>
                    <Button onClick={handleSearch}>搜索</Button>
                  </div>
                </div>
              </div>

              {/* 表格 */}
              <div className="flex-1 overflow-auto p-6">
                <DocumentDataTable
                  data={data}
                  loading={loading}
                  page={page}
                  pageSize={pageSize}
                  total={total}
                  onPageChange={setPage}
                  onPageSizeChange={handlePageSizeChange}
                  onDownload={handleDownload}
                  onDelete={handleDelete}
                  onEditCanvas={(doc) =>
                    navigate({ to: "/wiki/tiptap-editor", search: { docId: doc.id } })
                  }
                />
              </div>
            </>
          )}
        </ResizablePanel>
      </ResizablePanelGroup>

      {/* 上传弹窗 */}
      <UploadDialog
        open={uploadOpen}
        defaultFolderId={activeKbId}
        onOpenChange={setUploadOpen}
        onSuccess={loadDocuments}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        title="删除文档"
        description={`确定要删除文档"${deleteTarget?.filename}"吗？`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
