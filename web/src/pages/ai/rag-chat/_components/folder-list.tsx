import { IconFolder, IconLoader2 } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import type { Folder } from "@/api/wiki/folder";
import { folderApi } from "@/api/wiki/folder";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface FolderListProps {
  value?: string;
  onChange: (id: string) => void;
}

export function FolderList({ value, onChange }: FolderListProps) {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const loadFolders = async () => {
      setLoading(true);
      try {
        const list = await folderApi.list();
        const all = Array.isArray(list) ? list : [];
        setFolders(all);

        if (all.length > 0 && !value) {
          onChange(all[0]!.id);
        }
      } catch {
        toast.error("加载文件夹失败");
      } finally {
        setLoading(false);
      }
    };

    loadFolders();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <IconLoader2 className="mr-2 h-4 w-4 animate-spin" />
        <span className="text-sm text-muted-foreground">加载中...</span>
      </div>
    );
  }

  return (
    <div className={cn("h-full overflow-auto pb-12 p-4 space-y-1")}>
      {folders.length === 0 ? (
        <div className="flex items-center justify-center p-8">
          <span className="text-sm text-muted-foreground">暂无文件夹</span>
        </div>
      ) : (
        folders.map((folder) => (
          <Button
            key={folder.id}
            variant={value === folder.id ? "default" : "ghost"}
            className="w-full justify-start"
            onClick={() => onChange(folder.id)}
          >
            <IconFolder className="mr-2 h-4 w-4" />
            {folder.name}
          </Button>
        ))
      )}
    </div>
  );
}
