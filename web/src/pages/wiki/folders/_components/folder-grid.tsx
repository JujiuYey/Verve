import { IconFolder } from "@tabler/icons-react";

import type { Folder } from "@/api/wiki/folder";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

import { FolderCard } from "./folder-card";

interface FolderGridProps {
  data: Folder[];
  loading?: boolean;
  onEdit: (folder: Folder) => void;
  onDelete: (folder: Folder) => void;
  onEnter?: (folder: Folder) => void;
}

export function FolderGrid({
  data,
  loading,
  onEdit,
  onDelete,
  onEnter,
}: FolderGridProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="h-32 animate-pulse rounded-lg border bg-muted" />
        ))}
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <Empty className="border border-dashed">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <IconFolder />
          </EmptyMedia>
          <EmptyTitle>暂无文件夹</EmptyTitle>
          <EmptyDescription>创建一个新文件夹开始使用</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {data.map((folder) => (
        <FolderCard
          key={folder.id}
          folder={folder}
          onEdit={onEdit}
          onDelete={onDelete}
          onEnter={onEnter}
        />
      ))}
    </div>
  );
}
