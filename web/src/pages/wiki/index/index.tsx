import { IconPlus, IconRefresh, IconSearch, IconUpload } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { type IndexJobProgress, ragWikiApi } from "@/api/rag/wiki";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import { BreadcrumbNav } from "./_components/breadcrumb-nav";
import { FolderFormModal } from "./_components/folder-form-modal";
import { ItemGrid } from "./_components/item-grid";
import { UploadDialog } from "./_components/upload-dialog";
import { getFolderContentView } from "./_shared/content-view";

const sortFolders = (folders: Folder[]) =>
  [...folders].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order;
    return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
  });

export function WikiIndexPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Folder[]>([]);
  const [folderTreeData, setFolderTreeData] = useState<FolderTreeNode[]>([]);
  const [loading, setLoading] = useState(false);

  // 文档相关状态
  const [documents, setDocuments] = useState<Document[]>([]);
  const [indexJobsByDocumentId, setIndexJobsByDocumentId] = useState<
    Record<string, IndexJobProgress | undefined>
  >({});
  const [documentsLoading, setDocumentsLoading] = useState(false);
  const [deleteDocumentTarget, setDeleteDocumentTarget] = useState<Document | null>(null);
  // const [deletingDocument, setDeletingDocument] = useState(false);

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

  const loadIndexJobs = useCallback(async () => {
    try {
      const jobs = await ragWikiApi.listJobs();
      const nextJobsByDocumentId: Record<string, IndexJobProgress> = {};
      for (const job of jobs) {
        if (!nextJobsByDocumentId[job.document_id]) {
          nextJobsByDocumentId[job.document_id] = job;
        }
      }
      setIndexJobsByDocumentId(nextJobsByDocumentId);
    } catch {
      toast.error("加载文档解析状态失败");
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
      void loadIndexJobs();
    } else {
      setDocuments([]);
    }
  }, [currentFolderId, loadDocuments, loadIndexJobs]);

  useEffect(() => {
    if (!currentFolderId || documents.length === 0) return;

    const hasActiveJob = documents.some((document) => {
      const status = indexJobsByDocumentId[document.id]?.status;
      return status === "pending" || status === "running";
    });
    if (!hasActiveJob) return;

    const intervalId = window.setInterval(() => {
      void loadIndexJobs();
    }, 4000);

    return () => window.clearInterval(intervalId);
  }, [currentFolderId, documents, indexJobsByDocumentId, loadIndexJobs]);

  // 删除文档
  const handleDeleteDocument = useCallback((doc: Document) => {
    setDeleteDocumentTarget(doc);
  }, []);

  const handleOpenDocument = useCallback(
    (doc: Document) => {
      navigate({
        to: "/learn/feynman-practice/$documentId",
        params: { documentId: doc.id },
      });
    },
    [navigate],
  );

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
        setData((prev) => sortFolders([...prev, createdFolder]));
        toast.success("创建成功");
      } else {
        const updatedFolder = await folderApi.update(formData as UpdateFolderRequest);
        setData((prev) =>
          sortFolders(
            prev.map((folder) => (folder.id === updatedFolder.id ? updatedFolder : folder)),
          ),
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
      <div
        key={currentFolder ? "with-detail" : "without-detail"}
        className="min-h-0 flex-1 overflow-hidden"
      >
        <div id="folder-content-panel" className="h-full min-w-0 overflow-y-auto">
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

            <div className="flex-1 min-h-0">
              <div className="min-h-0 min-w-0">
                <ItemGrid
                  folders={contentView.folders}
                  documents={contentView.documents}
                  indexJobsByDocumentId={indexJobsByDocumentId}
                  loading={loading || documentsLoading}
                  onEditFolder={handleEdit}
                  onDeleteFolder={handleDelete}
                  onEnterFolder={handleEnterFolder}
                  onDeleteDocument={handleDeleteDocument}
                  onOpenDocument={(document) => void handleOpenDocument(document)}
                  onIndexStatusRefresh={() => void loadIndexJobs()}
                />
              </div>
            </div>
          </div>
        </div>
      </div>

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
            setIndexJobsByDocumentId((prev) => ({
              ...prev,
              [document.id]: {
                id: `local-${document.id}`,
                document_id: document.id,
                status: "pending",
                chunk_count: 0,
                created_at: new Date().toISOString(),
              },
            }));
            window.setTimeout(() => void loadIndexJobs(), 1000);
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
