import { IconPencil, IconShare, IconUserPlus } from "@tabler/icons-react";

import type { Folder } from "@/api/wiki/folder";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

import { folderIconAsset } from "../_shared/file-icons";

interface FolderDetailPanelProps {
  folder: Folder;
  onEdit: (folder: Folder) => void;
}

export function FolderDetailPanel({ folder, onEdit }: FolderDetailPanelProps) {
  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-col gap-4 p-4">
      {/* 标题栏 */}
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold truncate">{folder.name}</h3>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 shrink-0"
          onClick={() => onEdit(folder)}
        >
          <IconPencil className="h-4 w-4" />
        </Button>
      </div>

      {/* 文件夹图标预览 */}
      <div className="flex items-center justify-center rounded-lg border bg-muted/30 h-40">
        <img
          src={folderIconAsset.src}
          alt={folderIconAsset.alt}
          className="h-20 w-20 object-contain"
        />
      </div>

      <Button variant="outline" className="w-full">
        <IconShare className="mr-2 h-4 w-4" />
        分享
      </Button>

      <Separator />

      {/* 协作者 */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm text-muted-foreground">协作者</span>
          <Button variant="ghost" size="icon" className="h-7 w-7">
            <IconUserPlus className="h-4 w-4 text-muted-foreground" />
          </Button>
        </div>
        <div className="flex items-center gap-2">
          {folder.created_by && (
            <Avatar className="h-8 w-8">
              <AvatarFallback className="text-xs">{folder.created_by.slice(0, 1)}</AvatarFallback>
            </Avatar>
          )}
        </div>
      </div>

      <Separator />

      {/* 描述 */}
      <div>
        <span className="text-sm text-muted-foreground">描述</span>
        <p className="mt-1 text-sm font-medium">{folder.description || "暂无描述"}</p>
      </div>

      <Separator />

      {/* 所有者 */}
      <div>
        <span className="text-sm text-muted-foreground">所有者</span>
        <p className="mt-1 text-sm font-medium truncate">
          {folder.created_by_user?.full_name || "未知"}
        </p>
      </div>
    </div>
  );
}
