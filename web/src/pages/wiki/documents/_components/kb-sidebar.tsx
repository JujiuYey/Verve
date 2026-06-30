import { IconFolder } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { type Folder, folderApi } from "@/api/wiki/folder";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface KbSidebarProps {
  value?: string;
  onChange: (id: string) => void;
  className?: string;
}

export function KbSidebar({ value, onChange, className }: KbSidebarProps) {
  const [loading, setLoading] = useState(false);
  const [folders, setFolders] = useState<Folder[]>([]);

  const loadFolders = async () => {
    setLoading(true);
    try {
      const list = await folderApi.list();
      const all = list || [];
      setFolders(all);

      if (all.length > 0 && !value) {
        onChange(all[0]!.id);
      }
    } catch {
      toast.error("加载文件夹失败", {
        description: "获取文件夹列表失败",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFolders();
  }, []);

  return (
    <div className={cn("h-full overflow-auto pb-12", className)}>
      {loading ? (
        <div className="flex items-center justify-center p-8">
          <div className="text-sm text-muted-foreground">加载中...</div>
        </div>
      ) : folders.length === 0 ? (
        <div className="flex items-center justify-center p-8">
          <div className="text-sm text-muted-foreground">暂无文件夹</div>
        </div>
      ) : (
        <div className="space-y-1 p-4">
          {folders.map((folder) => (
            <Button
              key={folder.id}
              variant={value === folder.id ? "default" : "ghost"}
              className="w-full justify-start"
              onClick={() => onChange(folder.id)}
            >
              <IconFolder className="mr-2 h-4 w-4" />
              {folder.name}
            </Button>
          ))}
        </div>
      )}
    </div>
  );
}
