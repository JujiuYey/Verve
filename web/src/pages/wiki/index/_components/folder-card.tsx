import { IconDotsVertical, IconPencil, IconRefresh, IconTrash } from "@tabler/icons-react";
import { useState } from "react";

import type { Folder } from "@/api/wiki/folder";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { folderIconAsset } from "../_shared/file-icons";

export interface FolderCardProps {
  folder: Folder;
  onEdit: (folder: Folder) => void;
  onDelete: (folder: Folder) => void;
  onEnter?: (folder: Folder) => void;
  onIndex?: (folder: Folder) => void;
  indexing?: boolean;
}

export function FolderCard({ folder, onEdit, onDelete, onEnter, onIndex, indexing }: FolderCardProps) {
  const [menuOpen, setMenuOpen] = useState(false);

  const handleClick = () => {
    onEnter?.(folder);
  };

  const handleMenuClick = (e: React.MouseEvent) => {
    e.stopPropagation();
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    onEdit(folder);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(folder);
  };

  const handleIndex = (e: React.MouseEvent) => {
    e.stopPropagation();
    onIndex?.(folder);
    setMenuOpen(false);
  };

  return (
    <div
      className="group relative rounded-lg border bg-card p-4 transition-all hover:border-primary hover:shadow-md cursor-pointer"
      onClick={handleClick}
    >
      <div className="flex items-center gap-3 pr-8">
        <div className="rounded-md bg-muted p-2">
          <img
            src={folderIconAsset.src}
            alt={folderIconAsset.alt}
            className="h-6 w-6 object-contain"
          />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-medium truncate">{folder.name}</h3>
          {folder.description && (
            <p className="text-sm text-muted-foreground line-clamp-2 mt-1">{folder.description}</p>
          )}
        </div>
      </div>
      <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="secondary"
            size="icon"
            className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8"
            onClick={handleMenuClick}
          >
            <IconDotsVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {onIndex && (
            <>
              <DropdownMenuItem onClick={handleIndex} disabled={indexing}>
                <IconRefresh className="mr-2 h-4 w-4" />
                {indexing ? "解析中..." : "解析知识库"}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
            </>
          )}
          <DropdownMenuItem onClick={handleEdit}>
            <IconPencil className="mr-2 h-4 w-4" />
            编辑
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
