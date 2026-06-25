import { IconCheck, IconEdit, IconTrash, IconX } from "@tabler/icons-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

interface RagSession {
  id: string;
  title?: string;
  folder_id: string;
  document_id?: string;
  message_count: number;
  total_tokens?: number;
  created_at: string;
  updated_at: string;
}

interface SessionListProps {
  items: RagSession[];
  selectedId?: string;
  onSelect: (id: string) => void;
  onDeleted: () => void;
  onTitleUpdated: () => void;
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "刚刚";
  if (diffMin < 60) return `${diffMin} 分钟前`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour} 小时前`;
  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 30) return `${diffDay} 天前`;
  const diffMonth = Math.floor(diffDay / 30);
  if (diffMonth < 12) return `${diffMonth} 个月前`;
  return `${Math.floor(diffMonth / 12)} 年前`;
}

function formatTokenCount(tokens: number): string {
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M tokens`;
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}k tokens`;
  return `${tokens} tokens`;
}

export function SessionList({
  items,
  selectedId,
  onSelect,
  onDeleted,
  onTitleUpdated,
}: SessionListProps) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState("");
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleStartEdit = (e: React.MouseEvent, item: RagSession) => {
    e.stopPropagation();
    setEditingId(item.id);
    setEditTitle(item.title || "");
    setTimeout(() => inputRef.current?.focus(), 0);
  };

  const handleSaveTitle = async (e: React.MouseEvent) => {
    e.stopPropagation();
    const trimmed = editTitle.trim();
    if (!trimmed) {
      setEditingId(null);
      return;
    }
    try {
      // Mock: 模拟更新标题
      toast.success("标题已更新");
      onTitleUpdated();
    } catch {
      toast.error("更新标题失败");
    }
    setEditingId(null);
  };

  const handleCancelEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    setEditingId(null);
  };

  const handleDelete = (e: React.MouseEvent, sessionId: string) => {
    e.stopPropagation();
    setDeleteTargetId(sessionId);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTargetId) return;
    try {
      // Mock: 模拟删除
      toast.success("对话已删除");
      onDeleted();
    } catch {
      toast.error("删除对话失败");
    }
  };

  return (
    <div className="flex-1 overflow-auto" style={{ height: "calc(100vh - 190px)" }}>
      <div className="flex flex-col gap-2 p-4 pt-0">
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            className={cn(
              "flex flex-col items-start gap-2 rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent group",
              selectedId === item.id && "bg-muted",
            )}
            onClick={() => onSelect(item.id)}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                {editingId === item.id ? (
                  <div
                    className="flex items-center gap-1 flex-1"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <Input
                      ref={inputRef}
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                      className="h-6 text-sm"
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleSaveTitle(e as unknown as React.MouseEvent);
                        if (e.key === "Escape") setEditingId(null);
                      }}
                    />
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={handleSaveTitle}
                    >
                      <IconCheck className="h-3 w-3" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={handleCancelEdit}
                    >
                      <IconX className="h-3 w-3" />
                    </Button>
                  </div>
                ) : (
                  <>
                    <div className="font-semibold flex-1 truncate">
                      {item.title || "未命名对话"}
                    </div>
                    <div className="ml-auto flex items-center gap-1">
                      <span
                        className={cn(
                          "text-xs",
                          selectedId === item.id ? "text-foreground" : "text-muted-foreground",
                        )}
                      >
                        {formatRelativeTime(item.updated_at)}
                      </span>
                      <div className="hidden group-hover:flex items-center gap-0.5 ml-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6"
                          onClick={(e) => handleStartEdit(e, item)}
                        >
                          <IconEdit className="h-3 w-3" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6 text-destructive"
                          onClick={(e) => handleDelete(e, item.id)}
                        >
                          <IconTrash className="h-3 w-3" />
                        </Button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </div>
            <div className="line-clamp-2 text-xs text-muted-foreground">
              {item.message_count} 条消息
              {item.total_tokens != null && (
                <span className="ml-2">
                  &middot;
                  {formatTokenCount(item.total_tokens)}
                </span>
              )}
            </div>
          </button>
        ))}
        {items.length === 0 && (
          <div className="flex items-center justify-center p-8 text-sm text-muted-foreground">
            暂无对话
          </div>
        )}
      </div>
      <ConfirmDialog
        open={!!deleteTargetId}
        title="删除对话"
        description="确定要删除此对话吗？"
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTargetId(null);
        }}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
