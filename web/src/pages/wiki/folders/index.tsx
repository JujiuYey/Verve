import { IconPlus, IconRefresh, IconSearch, IconUpload } from "@tabler/icons-react";
// import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import type { Document } from "@/api/wiki/document";
import { documentApi } from "@/api/wiki/document";
import {
  type CreateFolderRequest,
  type Folder,
  folderApi,
  type FolderTreeNode,
  type UpdateFolderRequest,
} from "@/api/wiki/folder";
import { ConfirmDialog } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";

import { BreadcrumbNav } from "./_components/breadcrumb-nav";
import { FolderDetailPanel } from "./_components/folder-detail-panel";
import { FolderFormModal } from "./_components/folder-form-modal";
import { FolderTree } from "./_components/folder-tree";
import { ItemGrid } from "./_components/item-grid";
import { UploadDialog } from "./_components/upload-dialog";
import { getFolderContentView } from "./_shared/content-view";

export function FoldersPage() {
  // const navigate = useNavigate();
  const [data, setData] = useState<Folder[]>([]);
  const [folderTreeData, setFolderTreeData] = useState<FolderTreeNode[]>([]);
  const [loading, setLoading] = useState(false);

  // 文档相关状态
  const [documents, setDocuments] = useState<Document[]>([]);
  const [documentsLoading, setDocumentsLoading] = useState(false);
  const [deleteDocumentTarget, setDeleteDocumentTarget] = useState<Document | null>(null);
  // const [deletingDocument, setDeletingDocument] = useState(false);

  // Tab 切换状态
  const [activeTab, setActiveTab] = useState<"all" | "folders" | "documents">("all");

  // 面包屑导航状态
  const [breadcrumb, setBreadcrumb] = useState<{ id?: string; name: string }[]>([]);
  // 当前进入的文件夹对象（用于右侧详情面板）
  const [currentFolder, setCurrentFolder] = useState<Folder | null>(null);

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState("");

  const [uploadOpen, setUploadOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null);

  // 获取当前文件夹ID（面包屑最后一个）
  const currentFolderId = breadcrumb.length > 0 ? breadcrumb[breadcrumb.length - 1].id : undefined;

  const loadFolders = useCallback(async (parentId?: string) => {
    setLoading(true);
    try {
      const res = await folderApi.list(parentId);
      setData(res || []);
    } catch {
      toast.error("加载文件夹列表失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const loadFolderTree = useCallback(async () => {
    try {
      const res = await folderApi.tree();
      setFolderTreeData(res || []);
    } catch {
      toast.error("加载文件夹树失败");
    }
  }, []);

  // 加载文档列表
  const loadDocuments = useCallback(async (folderId: string) => {
    setDocumentsLoading(true);
    try {
      const res = await documentApi.list({ folder_id: folderId });
      setDocuments(res || []);
    } catch {
      toast.error("加载文档列表失败");
    } finally {
      setDocumentsLoading(false);
    }
  }, []);

  // 加载文件夹列表，使用当前面包屑路径的最后一个文件夹ID
  useEffect(() => {
    void loadFolders(currentFolderId);
  }, [loadFolders, currentFolderId]);

  // 当进入文件夹时，加载该文件夹下的文档
  useEffect(() => {
    if (currentFolderId) {
      void loadDocuments(currentFolderId);
    } else {
      setDocuments([]);
    }
  }, [currentFolderId, loadDocuments]);

  // 删除文档
  const handleDeleteDocument = useCallback((doc: Document) => {
    setDeleteDocumentTarget(doc);
  }, []);

  const handleConfirmDeleteDocument = async () => {
    if (!deleteDocumentTarget) return;
    // setDeletingDocument(true);
    try {
      await documentApi.delete(deleteDocumentTarget.id);
      toast.success("删除成功");
      setDeleteDocumentTarget(null);
      if (currentFolderId) {
        void loadDocuments(currentFolderId);
      }
    } catch (error) {
      toast.error(`删除失败，${error}`);
    }
    // finally {
    //   setDeletingDocument(false);
    // }
  };

  const handleRefresh = useCallback(() => {
    void loadFolders(currentFolderId);
    if (currentFolderId) {
      void loadDocuments(currentFolderId);
    }
  }, [currentFolderId, loadDocuments, loadFolders]);

  useEffect(() => {
    void loadFolderTree();
  }, [loadFolderTree]);

  // 进入文件夹
  const handleEnterFolder = useCallback((folder: Folder) => {
    setBreadcrumb((prev) => [...prev, { id: folder.id, name: folder.name }]);
    setCurrentFolder(folder);
  }, []);

  // 面包屑导航点击
  const handleBreadcrumbNavigate = useCallback(
    (item: { id?: string; name: string } | null, index: number) => {
      if (item === null) {
        // 返回根目录
        setBreadcrumb([]);
        setCurrentFolder(null);
      } else {
        // 导航到指定层级
        setBreadcrumb((prev) => prev.slice(0, index + 1));
        // 面包屑导航回退时重新获取文件夹详情
        if (item.id) {
          folderApi
            .findOne(item.id)
            .then(setCurrentFolder)
            .catch((error) => {
              console.error("获取文件夹详情失败:", error);
            });
        }
      }
    },
    [],
  );

  // 树形导航选择文件夹
  const handleTreeSelect = useCallback(
    async (folder: FolderTreeNode | null) => {
      if (folder === null) {
        // 返回根目录
        setBreadcrumb([]);
        setCurrentFolder(null);
        return;
      }

      // 检查是否在当前面包屑路径中
      const existingIndex = breadcrumb.findIndex((item) => item.id === folder.id);
      if (existingIndex !== -1) {
        // 已在路径中，导航到该位置
        handleBreadcrumbNavigate({ id: folder.id, name: folder.name }, existingIndex);
      } else {
        // 不在路径中，进入该文件夹
        setBreadcrumb([{ id: folder.id, name: folder.name }]);
        // 获取完整文件夹信息用于右侧面板
        try {
          const fullFolder = await folderApi.findOne(folder.id);
          setCurrentFolder(fullFolder);
        } catch (error) {
          console.error("获取文件夹详情失败:", error);
        }
      }
    },
    [breadcrumb, handleBreadcrumbNavigate],
  );

  const handleCreate = () => {
    setFormMode("create");
    setSelectedFolder(null);
    setFormOpen(true);
  };

  const handleEdit = (folder: Folder) => {
    setFormMode("edit");
    setSelectedFolder(folder);
    setFormOpen(true);
  };

  const handleDelete = (folder: Folder) => {
    setDeleteTarget(folder);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await folderApi.delete(deleteTarget.id);
      toast.success("删除成功");
      void loadFolders(currentFolderId);
      void loadFolderTree();
    } catch (error) {
      toast.error(`删除失败，${error}`);
    }
  };

  const handleSubmit = async (formData: CreateFolderRequest | UpdateFolderRequest) => {
    setSubmitting(true);
    try {
      if (formMode === "create") {
        // 创建时传入当前目录的 parent_id
        const createdFolder = await folderApi.create({
          ...(formData as CreateFolderRequest),
          parent_id: currentFolderId,
        });
        setData((prev) => [...prev, createdFolder]);
        toast.success("创建成功");
      } else {
        const updatedFolder = await folderApi.update(formData as UpdateFolderRequest);
        setData((prev) =>
          prev.map((folder) => (folder.id === updatedFolder.id ? updatedFolder : folder)),
        );
        setCurrentFolder((prev) => (prev?.id === updatedFolder.id ? updatedFolder : prev));
        toast.success("更新成功");
      }
      setFormOpen(false);
      void loadFolderTree();
    } catch (error) {
      console.error("保存文件夹失败:", error);
      toast.error(formMode === "create" ? "创建失败" : "更新失败");
    } finally {
      setSubmitting(false);
    }
  };

  const contentView = getFolderContentView({
    folders: data,
    documents,
    searchKeyword,
  });

  return (
    <div className="flex h-full min-h-0 flex-col">
      <ResizablePanelGroup
        key={currentFolder ? "with-detail" : "without-detail"}
        orientation="horizontal"
        className="min-h-0 flex-1 overflow-hidden border rounded-xl"
      >
        {/* 左侧文件夹树形导航 */}
        <ResizablePanel
          id="folder-tree-panel"
          defaultSize="20%"
          minSize="15%"
          maxSize="30%"
          className="min-w-0 overflow-y-auto"
        >
          <FolderTree
            data={folderTreeData}
            selectedId={currentFolderId}
            onSelect={handleTreeSelect}
          />
        </ResizablePanel>

        <ResizableHandle withHandle />

        <ResizablePanel
          id="folder-content-panel"
          defaultSize={currentFolder ? "65%" : "80%"}
          minSize="45%"
          className="min-w-0 overflow-y-auto"
        >
          <div className="flex h-full min-h-0 flex-col gap-4 p-2">
            <div className="flex items-center justify-between gap-4">
              <div className="relative max-w-sm flex-1">
                <IconSearch className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="搜索文件夹和文档..."
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                  className="pl-9"
                />
              </div>
              <div className="flex items-center gap-2">
                {currentFolderId && (
                  <>
                    <Button size="sm" variant="outline" onClick={handleRefresh}>
                      <IconRefresh className="h-4 w-4" />
                    </Button>
                    <Button onClick={() => setUploadOpen(true)}>
                      <IconUpload className="mr-2 h-4 w-4" />
                      上传文档
                    </Button>
                  </>
                )}
                <Button onClick={handleCreate}>
                  <IconPlus className="mr-2 h-4 w-4" />
                  添加文件夹
                </Button>
              </div>
            </div>

            {/* 面包屑导航 */}
            {breadcrumb.length > 0 && (
              <BreadcrumbNav items={breadcrumb} onNavigate={handleBreadcrumbNavigate} />
            )}

            {/* Tab 切换 */}
            <div className="flex items-center gap-1 border-b">
              <button
                className={`px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === "all"
                    ? "border-b-2 border-primary text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
                onClick={() => setActiveTab("all")}
              >
                全部
                <Badge variant="secondary" className="ml-2">
                  {contentView.counts.all}
                </Badge>
              </button>
              <button
                className={`px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === "folders"
                    ? "border-b-2 border-primary text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
                onClick={() => setActiveTab("folders")}
              >
                文件夹
                <Badge variant="secondary" className="ml-2">
                  {contentView.counts.folders}
                </Badge>
              </button>
              <button
                className={`px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === "documents"
                    ? "border-b-2 border-primary text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
                onClick={() => setActiveTab("documents")}
              >
                文档
                <Badge variant="secondary" className="ml-2">
                  {contentView.counts.documents}
                </Badge>
              </button>
            </div>

            <ItemGrid
              folders={contentView.folders}
              documents={contentView.documents}
              loading={loading || documentsLoading}
              activeTab={activeTab}
              onEditFolder={handleEdit}
              onDeleteFolder={handleDelete}
              onEnterFolder={handleEnterFolder}
              onDeleteDocument={handleDeleteDocument}
            />
          </div>
        </ResizablePanel>

        {currentFolder && (
          <>
            <ResizableHandle withHandle />
            <ResizablePanel
              id="folder-detail-panel"
              defaultSize="25%"
              minSize="20%"
              maxSize="35%"
              className="min-w-0 overflow-y-auto bg-background"
            >
              <FolderDetailPanel folder={currentFolder} onEdit={handleEdit} />
            </ResizablePanel>
          </>
        )}
      </ResizablePanelGroup>

      <FolderFormModal
        open={formOpen}
        mode={formMode}
        folder={selectedFolder}
        loading={submitting}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
      />

      <UploadDialog
        open={uploadOpen}
        defaultFolderId={currentFolderId}
        folderTree={folderTreeData}
        onOpenChange={setUploadOpen}
        onSuccess={(document, folderId) => {
          if (folderId === currentFolderId) {
            setDocuments((prev) => [...prev, { ...document, folder_id: folderId }]);
          }
        }}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        title="删除文件夹"
        description={`确定要删除文件夹"${deleteTarget?.name}"吗？此操作将删除该文件夹及其所有文档，且无法撤销。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleConfirmDelete}
      />

      {/* 删除文档确认对话框 */}
      <ConfirmDialog
        open={!!deleteDocumentTarget}
        title="删除文档"
        description={`确定要删除文档"${deleteDocumentTarget?.filename}"吗？此操作无法撤销。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteDocumentTarget(null);
        }}
        onConfirm={handleConfirmDeleteDocument}
      />
    </div>
  );
}
