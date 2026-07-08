import { IconFile, IconFolder } from "@tabler/icons-react";

import type { Document } from "@/api/wiki/document";
import type { Folder } from "@/api/wiki/folder";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

import { DocumentCard } from "./document-card";
import { FolderCard } from "./folder-card";

interface ItemGridProps {
  folders: Folder[];
  documents: Document[];
  loading?: boolean;
  onEditFolder: (folder: Folder) => void;
  onDeleteFolder: (folder: Folder) => void;
  onEnterFolder: (folder: Folder) => void;
  onDeleteDocument: (document: Document) => void;
  onOpenDocument?: (document: Document) => void;
  openingDocumentId?: string;
}

export function ItemGrid({
  folders,
  documents,
  loading,
  onEditFolder,
  onDeleteFolder,
  onEnterFolder,
  onDeleteDocument,
  onOpenDocument,
  openingDocumentId,
}: ItemGridProps) {
  const folderGridClassName = "grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4";
  const documentGridClassName = "grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3";

  const isEmpty = folders.length === 0 && documents.length === 0;

  if (loading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="h-32 animate-pulse rounded-lg border bg-muted" />
        ))}
      </div>
    );
  }

  if (isEmpty) {
    return (
      <Empty className="border border-dashed">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <IconFolder />
          </EmptyMedia>
          <EmptyTitle>暂无内容</EmptyTitle>
          <EmptyDescription>创建文件夹或上传文档开始使用</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="space-y-6">
      {/* 文件夹区域 */}
      {folders.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-2">
            <IconFolder className="h-4 w-4" />
            文件夹 ({folders.length})
          </h3>
          <div className={folderGridClassName}>
            {folders.map((folder) => (
              <FolderCard
                key={folder.id}
                folder={folder}
                onEdit={onEditFolder}
                onDelete={onDeleteFolder}
                onEnter={onEnterFolder}
              />
            ))}
          </div>
        </div>
      )}

      {/* 文档区域 */}
      {documents.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-2">
            <IconFile className="h-4 w-4" />
            文档 ({documents.length})
          </h3>
          <div className={documentGridClassName}>
            {documents.map((doc) => (
              <DocumentCard
                key={doc.id}
                document={doc}
                onDelete={onDeleteDocument}
                onOpen={onOpenDocument}
                opening={openingDocumentId === doc.id}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
