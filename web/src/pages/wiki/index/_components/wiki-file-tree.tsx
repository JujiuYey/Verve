import {
  ChevronRightIcon,
  FileTextIcon,
  FolderIcon,
  FolderOpenIcon,
  LibraryBigIcon,
  MoreHorizontalIcon,
  PencilIcon,
  Trash2Icon,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import type { Document } from "@/api/wiki/document";
import type { Folder, FolderTreeNode } from "@/api/wiki/folder";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

type FolderBranch = {
  folder: FolderTreeNode;
  folders: FolderBranch[];
  documents: Document[];
};

interface WikiFileTreeProps {
  folders: FolderTreeNode[];
  documents: Document[];
  loading: boolean;
  searchKeyword: string;
  selectedFolderId?: string;
  selectedDocumentId?: string;
  onSelectRoot: () => void;
  onSelectFolder: (folder: FolderTreeNode) => void;
  onSelectDocument: (document: Document) => void;
  onEditFolder: (folder: Folder) => void;
  onDeleteFolder: (folder: Folder) => void;
}

const sortDocuments = (documents: Document[]) =>
  [...documents].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order;
    return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
  });

function buildFolderBranches(
  folders: FolderTreeNode[],
  documentsByFolder: Map<string, Document[]>,
  keyword: string,
): FolderBranch[] {
  const build = (folder: FolderTreeNode, activeKeyword: string): FolderBranch | null => {
    const folderDocuments = sortDocuments(documentsByFolder.get(folder.id) ?? []);
    const folderMatches = folder.name.toLowerCase().includes(activeKeyword);

    if (!activeKeyword || folderMatches) {
      return {
        folder,
        folders: folder.children
          .map((child) => build(child, ""))
          .filter((child): child is FolderBranch => child !== null),
        documents: folderDocuments,
      };
    }

    const matchingFolders = folder.children
      .map((child) => build(child, activeKeyword))
      .filter((child): child is FolderBranch => child !== null);
    const matchingDocuments = folderDocuments.filter((document) =>
      document.filename.toLowerCase().includes(activeKeyword),
    );

    if (matchingFolders.length === 0 && matchingDocuments.length === 0) return null;

    return {
      folder,
      folders: matchingFolders,
      documents: matchingDocuments,
    };
  };

  return folders
    .map((folder) => build(folder, keyword))
    .filter((folder): folder is FolderBranch => folder !== null);
}

function collectFolderIds(branches: FolderBranch[]) {
  const ids = new Set<string>();
  const collect = (branch: FolderBranch) => {
    ids.add(branch.folder.id);
    branch.folders.forEach(collect);
  };
  branches.forEach(collect);
  return ids;
}

function toFolder(folder: FolderTreeNode): Folder {
  return {
    id: folder.id,
    name: folder.name,
    description: folder.description,
    parent_id: folder.parent_id,
    sort_order: folder.sort_order,
    created_at: folder.created_at,
    updated_at: folder.updated_at,
  };
}

export function WikiFileTree({
  folders,
  documents,
  loading,
  searchKeyword,
  selectedFolderId,
  selectedDocumentId,
  onSelectRoot,
  onSelectFolder,
  onSelectDocument,
  onEditFolder,
  onDeleteFolder,
}: WikiFileTreeProps) {
  const [expandedFolderIds, setExpandedFolderIds] = useState<Set<string>>(new Set());
  const normalizedKeyword = searchKeyword.trim().toLowerCase();

  const documentsByFolder = useMemo(() => {
    const grouped = new Map<string, Document[]>();
    documents.forEach((document) => {
      const current = grouped.get(document.folder_id) ?? [];
      current.push(document);
      grouped.set(document.folder_id, current);
    });
    return grouped;
  }, [documents]);

  const branches = useMemo(
    () => buildFolderBranches(folders, documentsByFolder, normalizedKeyword),
    [documentsByFolder, folders, normalizedKeyword],
  );

  useEffect(() => {
    if (folders.length === 0) return;
    setExpandedFolderIds((current) => {
      if (current.size > 0) return current;
      return new Set(folders.map((folder) => folder.id));
    });
  }, [folders]);

  const searchExpandedFolderIds = useMemo(
    () => (normalizedKeyword ? collectFolderIds(branches) : null),
    [branches, normalizedKeyword],
  );

  const toggleFolder = (folderId: string) => {
    setExpandedFolderIds((current) => {
      const next = new Set(current);
      if (next.has(folderId)) next.delete(folderId);
      else next.add(folderId);
      return next;
    });
  };

  const renderBranch = (branch: FolderBranch, level: number) => {
    const { folder } = branch;
    const hasChildren = branch.folders.length > 0 || branch.documents.length > 0;
    const isExpanded = searchExpandedFolderIds?.has(folder.id) ?? expandedFolderIds.has(folder.id);
    const isSelected = selectedFolderId === folder.id;

    return (
      <Collapsible key={folder.id} open={isExpanded}>
        <div
          className={cn(
            "group flex min-h-8 items-center rounded-md pr-1 transition-colors hover:bg-accent",
            isSelected && "bg-accent text-accent-foreground",
          )}
          style={{ paddingLeft: `${level * 14 + 4}px` }}
        >
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            className={cn("shrink-0", !hasChildren && "invisible")}
            aria-label={isExpanded ? `收起${folder.name}` : `展开${folder.name}`}
            onClick={() => toggleFolder(folder.id)}
          >
            <ChevronRightIcon className={cn("transition-transform", isExpanded && "rotate-90")} />
          </Button>

          <button
            type="button"
            className="flex min-w-0 flex-1 items-center gap-2 py-1.5 text-left text-sm outline-none focus-visible:underline"
            onClick={() => {
              onSelectFolder(folder);
              if (hasChildren) toggleFolder(folder.id);
            }}
          >
            {isExpanded ? (
              <FolderOpenIcon className="size-4 shrink-0 text-muted-foreground" />
            ) : (
              <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
            )}
            <span className="truncate" title={folder.name}>
              {folder.name}
            </span>
          </button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="shrink-0 opacity-0 group-focus-within:opacity-100 group-hover:opacity-100"
                aria-label={`${folder.name}操作`}
                onClick={(event) => event.stopPropagation()}
              >
                <MoreHorizontalIcon />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuGroup>
                <DropdownMenuItem onSelect={() => onEditFolder(toFolder(folder))}>
                  <PencilIcon />
                  编辑文件夹
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                <DropdownMenuItem
                  variant="destructive"
                  onSelect={() => onDeleteFolder(toFolder(folder))}
                >
                  <Trash2Icon />
                  删除文件夹
                </DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <CollapsibleContent>
          {branch.folders.map((child) => renderBranch(child, level + 1))}
          {branch.documents.map((document) => (
            <button
              key={document.id}
              type="button"
              className={cn(
                "flex min-h-8 w-full items-center gap-2 rounded-md py-1.5 pr-3 text-left text-sm text-muted-foreground outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring/50",
                selectedDocumentId === document.id &&
                  "bg-primary/10 font-medium text-primary hover:bg-primary/10 hover:text-primary",
              )}
              style={{ paddingLeft: `${(level + 1) * 14 + 34}px` }}
              onClick={() => onSelectDocument(document)}
            >
              <FileTextIcon className="size-4 shrink-0" />
              <span className="truncate" title={document.filename}>
                {document.filename}
              </span>
            </button>
          ))}
        </CollapsibleContent>
      </Collapsible>
    );
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-2 p-2">
        {Array.from({ length: 7 }).map((_, index) => (
          <Skeleton key={index} className={cn("h-8", index % 3 === 0 ? "w-4/5" : "w-full")} />
        ))}
      </div>
    );
  }

  return (
    <nav aria-label="知识库文件树" className="flex flex-col gap-1 p-2">
      <button
        type="button"
        className={cn(
          "flex min-h-9 items-center gap-2 rounded-md px-2 text-left text-sm font-medium outline-none transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring/50",
          !selectedFolderId && !selectedDocumentId && "bg-accent text-accent-foreground",
        )}
        onClick={onSelectRoot}
      >
        <LibraryBigIcon className="size-4 shrink-0 text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate">全部文档</span>
        <span className="text-xs font-normal text-muted-foreground">{documents.length}</span>
      </button>

      {branches.length > 0 ? (
        branches.map((branch) => renderBranch(branch, 0))
      ) : (
        <Empty className="min-h-56 border-0 px-4 py-8">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              {normalizedKeyword ? <FileTextIcon /> : <FolderIcon />}
            </EmptyMedia>
            <EmptyTitle className="text-sm">
              {normalizedKeyword ? "没有匹配内容" : "知识库还是空的"}
            </EmptyTitle>
            <EmptyDescription>
              {normalizedKeyword ? "换个关键词试试。" : "先创建文件夹，再上传 Markdown 文档。"}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </nav>
  );
}
