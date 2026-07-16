import { useNavigate } from "@tanstack/react-router";
import { FolderPlusIcon, RefreshCwIcon, SearchIcon, UploadIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import type { LearningAgentType } from "@/api/learning";
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
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useIsMobile } from "@/hooks/use-mobile";

import { DocumentReader } from "./_components/document-reader";
import { FolderFormModal } from "./_components/folder-form-modal";
import { UploadDialog } from "./_components/upload-dialog";
import { WikiFileTree } from "./_components/wiki-file-tree";

const practiceRoutes = {
  listener: "/learn/feynman-practice/$documentId/listener",
  teacher: "/learn/feynman-practice/$documentId/teacher",
  curator: "/learn/feynman-practice/$documentId/curator",
} as const satisfies Record<LearningAgentType, string>;

function flattenFolders(folders: FolderTreeNode[]): FolderTreeNode[] {
  return folders.flatMap((folder) => [folder, ...flattenFolders(folder.children)]);
}

export function WikiIndexPage() {
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const [folderTreeData, setFolderTreeData] = useState<FolderTreeNode[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [indexJobsByDocumentId, setIndexJobsByDocumentId] = useState<
    Record<string, IndexJobProgress | undefined>
  >({});
  const [loading, setLoading] = useState(true);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [selectedFolderId, setSelectedFolderId] = useState<string>();
  const [selectedDocumentId, setSelectedDocumentId] = useState<string>();

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [deleteFolderTarget, setDeleteFolderTarget] = useState<Folder | null>(null);
  const [deleteDocumentTarget, setDeleteDocumentTarget] = useState<Document | null>(null);

  const flatFolders = useMemo(() => flattenFolders(folderTreeData), [folderTreeData]);
  const activeFolder = useMemo(
    () => flatFolders.find((folder) => folder.id === selectedFolderId),
    [flatFolders, selectedFolderId],
  );
  const activeDocument = useMemo(
    () => documents.find((document) => document.id === selectedDocumentId),
    [documents, selectedDocumentId],
  );

  const loadLibrary = useCallback(async (showLoading = true) => {
    if (showLoading) setLoading(true);
    try {
      const [folders, allDocuments] = await Promise.all([folderApi.tree(), documentApi.list()]);
      setFolderTreeData(folders ?? []);
      setDocuments(allDocuments ?? []);
    } catch (error) {
      console.error("加载知识库失败:", error);
      toast.error("加载知识库失败");
    } finally {
      if (showLoading) setLoading(false);
    }
  }, []);

  const loadIndexJobs = useCallback(async () => {
    try {
      const jobs = await ragWikiApi.listJobs();
      const nextJobsByDocumentId: Record<string, IndexJobProgress> = {};
      jobs.forEach((job) => {
        if (!nextJobsByDocumentId[job.document_id]) {
          nextJobsByDocumentId[job.document_id] = job;
        }
      });
      setIndexJobsByDocumentId(nextJobsByDocumentId);
    } catch (error) {
      console.error("加载文档解析状态失败:", error);
      toast.error("加载文档解析状态失败");
    }
  }, []);

  useEffect(() => {
    void loadLibrary();
    void loadIndexJobs();
  }, [loadIndexJobs, loadLibrary]);

  useEffect(() => {
    const hasActiveIndexJob = documents.some((document) => {
      const status = indexJobsByDocumentId[document.id]?.status;
      return status === "pending" || status === "running";
    });
    if (!hasActiveIndexJob) return;

    const intervalId = window.setInterval(() => {
      void loadIndexJobs();
    }, 4000);
    return () => window.clearInterval(intervalId);
  }, [documents, indexJobsByDocumentId, loadIndexJobs]);

  useEffect(() => {
    if (selectedDocumentId && !documents.some((document) => document.id === selectedDocumentId)) {
      setSelectedDocumentId(undefined);
    }
  }, [documents, selectedDocumentId]);

  useEffect(() => {
    if (selectedFolderId && !flatFolders.some((folder) => folder.id === selectedFolderId)) {
      setSelectedFolderId(undefined);
    }
  }, [flatFolders, selectedFolderId]);

  const handleRefresh = useCallback(async () => {
    await Promise.all([loadLibrary(false), loadIndexJobs()]);
    toast.success("知识库已刷新");
  }, [loadIndexJobs, loadLibrary]);

  const handleCreateFolder = () => {
    setFormMode("create");
    setSelectedFolder(null);
    setFormOpen(true);
  };

  const handleEditFolder = (folder: Folder) => {
    setFormMode("edit");
    setSelectedFolder(folder);
    setFormOpen(true);
  };

  const handleSubmitFolder = async (formData: CreateFolderRequest | UpdateFolderRequest) => {
    setSubmitting(true);
    try {
      if (formMode === "create") {
        const createdFolder = await folderApi.create({
          ...(formData as CreateFolderRequest),
          parent_id: selectedFolderId,
        });
        setSelectedFolderId(createdFolder.id);
        setSelectedDocumentId(undefined);
        toast.success("文件夹已创建");
      } else {
        await folderApi.update(formData as UpdateFolderRequest);
        toast.success("文件夹已更新");
      }
      setFormOpen(false);
      await loadLibrary(false);
    } catch (error) {
      console.error("保存文件夹失败:", error);
      toast.error(formMode === "create" ? "创建文件夹失败" : "更新文件夹失败");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteFolder = async () => {
    if (!deleteFolderTarget) return;
    try {
      await folderApi.delete(deleteFolderTarget.id);
      setDeleteFolderTarget(null);
      setSelectedFolderId(undefined);
      setSelectedDocumentId(undefined);
      await loadLibrary(false);
      toast.success("文件夹已删除");
    } catch (error) {
      console.error("删除文件夹失败:", error);
      toast.error("删除文件夹失败");
    }
  };

  const handleDeleteDocument = async () => {
    if (!deleteDocumentTarget) return;
    try {
      await documentApi.delete(deleteDocumentTarget.id);
      setDocuments((current) =>
        current.filter((document) => document.id !== deleteDocumentTarget.id),
      );
      setIndexJobsByDocumentId((current) => {
        const next = { ...current };
        delete next[deleteDocumentTarget.id];
        return next;
      });
      if (selectedDocumentId === deleteDocumentTarget.id) setSelectedDocumentId(undefined);
      setDeleteDocumentTarget(null);
      toast.success("文档已删除");
    } catch (error) {
      console.error("删除文档失败:", error);
      toast.error("删除文档失败");
    }
  };

  return (
    <div className="h-full min-h-0 p-2">
      <ResizablePanelGroup
        key={isMobile ? "mobile" : "desktop"}
        orientation={isMobile ? "vertical" : "horizontal"}
        className="overflow-hidden rounded-lg border bg-background"
      >
        <ResizablePanel
          id="wiki-file-tree"
          defaultSize={isMobile ? "36%" : "280px"}
          minSize={isMobile ? "180px" : "200px"}
          maxSize={isMobile ? "55%" : "45%"}
          groupResizeBehavior={isMobile ? "preserve-relative-size" : "preserve-pixel-size"}
        >
          <aside className="flex h-full min-h-0 flex-col overflow-hidden bg-muted/10">
            <div className="flex shrink-0 flex-col gap-3 border-b p-3">
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <h1 className="text-sm font-semibold">知识库</h1>
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">
                    {activeFolder?.name ?? `${documents.length} 篇文档`}
                  </p>
                </div>

                <TooltipProvider>
                  <div className="flex shrink-0 items-center gap-1">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label="刷新知识库"
                          onClick={() => void handleRefresh()}
                        >
                          <RefreshCwIcon />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>刷新</TooltipContent>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label="上传文档"
                          disabled={flatFolders.length === 0}
                          onClick={() => setUploadOpen(true)}
                        >
                          <UploadIcon />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {flatFolders.length === 0 ? "请先创建文件夹" : "上传文档"}
                      </TooltipContent>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={
                            activeFolder ? `在${activeFolder.name}中新建文件夹` : "新建文件夹"
                          }
                          onClick={handleCreateFolder}
                        >
                          <FolderPlusIcon />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {activeFolder ? "新建子文件夹" : "新建文件夹"}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </TooltipProvider>
              </div>

              <InputGroup>
                <InputGroupAddon>
                  <SearchIcon />
                </InputGroupAddon>
                <InputGroupInput
                  value={searchKeyword}
                  placeholder="搜索文件夹和文档"
                  aria-label="搜索文件夹和文档"
                  onChange={(event) => setSearchKeyword(event.target.value)}
                />
              </InputGroup>
            </div>

            <ScrollArea className="min-h-0 flex-1">
              <WikiFileTree
                folders={folderTreeData}
                documents={documents}
                loading={loading}
                searchKeyword={searchKeyword}
                selectedFolderId={selectedDocumentId ? undefined : selectedFolderId}
                selectedDocumentId={selectedDocumentId}
                onSelectRoot={() => {
                  setSelectedFolderId(undefined);
                  setSelectedDocumentId(undefined);
                }}
                onSelectFolder={(folder) => {
                  setSelectedFolderId(folder.id);
                  setSelectedDocumentId(undefined);
                }}
                onSelectDocument={(document) => {
                  setSelectedFolderId(document.folder_id);
                  setSelectedDocumentId(document.id);
                }}
                onEditFolder={handleEditFolder}
                onDeleteFolder={setDeleteFolderTarget}
              />
            </ScrollArea>
          </aside>
        </ResizablePanel>

        <ResizableHandle withHandle />

        <ResizablePanel id="wiki-document-reader" minSize={isMobile ? "260px" : "320px"}>
          <main className="h-full min-h-0 overflow-hidden">
            <DocumentReader
              document={activeDocument}
              indexJob={activeDocument ? indexJobsByDocumentId[activeDocument.id] : undefined}
              onDelete={setDeleteDocumentTarget}
              onIndexStatusRefresh={() => void loadIndexJobs()}
              onOpenPractice={(document, agentType) => {
                void navigate({
                  to: practiceRoutes[agentType],
                  params: { documentId: document.id },
                });
              }}
            />
          </main>
        </ResizablePanel>
      </ResizablePanelGroup>

      <FolderFormModal
        open={formOpen}
        mode={formMode}
        folder={selectedFolder}
        loading={submitting}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmitFolder}
      />

      <UploadDialog
        open={uploadOpen}
        defaultFolderId={selectedFolderId}
        folderTree={folderTreeData}
        onOpenChange={setUploadOpen}
        onSuccess={(document, folderId) => {
          setDocuments((current) => [...current, { ...document, folder_id: folderId }]);
          setIndexJobsByDocumentId((current) => ({
            ...current,
            [document.id]: {
              id: `local-${document.id}`,
              document_id: document.id,
              document_version: document.current_version,
              status: "pending",
              chunk_count: 0,
              created_at: new Date().toISOString(),
            },
          }));
          setSelectedFolderId(folderId);
          setSelectedDocumentId((current) => current ?? document.id);
          window.setTimeout(() => void loadIndexJobs(), 1000);
        }}
      />

      <ConfirmDialog
        open={!!deleteFolderTarget}
        title="删除文件夹"
        description={`确定要删除文件夹“${deleteFolderTarget?.name}”吗？此操作会删除其子文件夹和文档，且无法撤销。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteFolderTarget(null);
        }}
        onConfirm={handleDeleteFolder}
      />

      <ConfirmDialog
        open={!!deleteDocumentTarget}
        title="删除文档"
        description={`确定要删除文档“${deleteDocumentTarget?.filename}”吗？此操作无法撤销。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteDocumentTarget(null);
        }}
        onConfirm={handleDeleteDocument}
      />
    </div>
  );
}
