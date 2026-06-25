import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

interface AgentSession {
  id: string;
  title?: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

interface SessionListProps {
  items: AgentSession[];
  selectedSession?: string;
  onSelect: (id: string) => void;
  loading?: boolean;
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diff = now - date;
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (seconds < 60) return "刚刚";
  if (minutes < 60) return `${minutes} 分钟前`;
  if (hours < 24) return `${hours} 小时前`;
  if (days < 30) return `${days} 天前`;
  return new Date(dateStr).toLocaleDateString("zh-CN");
}

export function SessionList({ items, selectedSession, onSelect, loading }: SessionListProps) {
  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner className="size-6" />
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 pt-0">
      <div className="flex flex-col gap-2">
        {items.map((item) => (
          <button
            key={item.id}
            className={cn(
              "flex flex-col items-start gap-2 rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent",
              selectedSession === item.id && "bg-muted",
            )}
            onClick={() => onSelect(item.id)}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <div className="font-semibold">{item.title || "未命名对话"}</div>
                </div>
                <div
                  className={cn(
                    "ml-auto text-xs",
                    selectedSession === item.id ? "text-foreground" : "text-muted-foreground",
                  )}
                >
                  {formatRelativeTime(item.updated_at)}
                </div>
              </div>
            </div>
            <div className="line-clamp-2 text-xs text-muted-foreground">
              {item.message_count} 条消息
            </div>
          </button>
        ))}
        {items.length === 0 && (
          <div className="py-8 text-center text-sm text-muted-foreground">暂无对话记录</div>
        )}
      </div>
    </div>
  );
}
