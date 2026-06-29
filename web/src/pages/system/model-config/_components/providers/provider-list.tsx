import { IconFilter, IconPlus, IconSearch, IconTrash } from "@tabler/icons-react";

import type { AIModel, AIPlatform } from "@/api/ai/model-config";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { getProviderLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

export type ProviderId = string;

interface ProviderListProps {
  platforms: AIPlatform[];
  enabledModels: AIModel[];
  selectedId: ProviderId | null;
  search: string;
  onSearchChange: (value: string) => void;
  onSelect: (id: ProviderId) => void;
  onAdd?: () => void;
  onDelete?: (id: ProviderId) => void;
}

function getInitials(name: string) {
  return name.slice(0, 2).toUpperCase();
}

function getPlatformState(configured: boolean, modelCount: number) {
  if (!configured) {
    return {
      label: "未接入",
      className: "bg-muted text-muted-foreground",
      description: "未配置平台密钥",
    };
  }
  if (modelCount === 0) {
    return {
      label: "已接入",
      className: "bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300",
      description: "已接入，未启用模型",
    };
  }
  return {
    label: "可用",
    className: "bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300",
    description: "已启用模型，可供业务调用",
  };
}

export function ProviderList({
  platforms,
  enabledModels,
  selectedId,
  search,
  onSearchChange,
  onSelect,
  onAdd,
  onDelete,
}: ProviderListProps) {
  const visiblePlatforms = platforms.filter((platform) => {
    const keyword = search.trim().toLowerCase();
    if (!keyword) return true;
    return `${platform.name} ${platform.id}`.toLowerCase().includes(keyword);
  });

  return (
    <aside className="flex h-full w-[300px] shrink-0 flex-col border-r bg-background">
      <div className="flex items-center border-b p-4">
        <div className="flex w-full items-center gap-2">
          <div className="relative flex-1">
            <IconSearch className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="model-platform-search"
              name="model-platform-search"
              type="search"
              value={search}
              onChange={(event) => onSearchChange(event.target.value)}
              placeholder="搜索模型平台..."
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              aria-label="搜索模型平台"
              className="pl-9"
            />
          </div>
          <Button size="icon">
            <IconFilter className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <nav className="space-y-1 p-3">
          {visiblePlatforms.map((platform) => {
            const isSelected = selectedId === platform.id;
            const modelCount = enabledModels.filter(
              (model) => model.platform_id === platform.id && model.status === "active",
            ).length;
            const state = getPlatformState(Boolean(platform.api_key_hint?.trim()), modelCount);
            const logo = getProviderLogo(platform);

            return (
              <div
                key={platform.id}
                className={cn(
                  "group relative flex items-center rounded-md transition-all",
                  isSelected
                    ? "bg-primary/10 text-primary shadow-xs ring-1 ring-primary/20"
                    : "hover:bg-muted/70 hover:ring-1 hover:ring-border/70",
                )}
              >
                <button
                  type="button"
                  className={cn(
                    "flex min-w-0 flex-1 items-center gap-3 rounded-md px-3 py-2.5 text-left outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
                    onDelete && "pr-12",
                  )}
                  onClick={() => onSelect(platform.id)}
                >
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-background shadow-xs ring-1 ring-border">
                    {logo ? (
                      <img
                        src={logo}
                        alt=""
                        className="h-6 w-6 rounded-sm object-contain"
                        draggable={false}
                      />
                    ) : (
                      <span className="text-[10px] font-bold text-muted-foreground">
                        {getInitials(platform.name)}
                      </span>
                    )}
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="flex min-w-0 items-center gap-2">
                      <p
                        className={cn(
                          "truncate text-sm font-semibold",
                          isSelected ? "text-primary" : "text-foreground",
                        )}
                      >
                        {platform.name}
                      </p>
                      <Badge
                        variant="secondary"
                        className={cn("h-5 shrink-0 px-1.5 text-[10px]", state.className)}
                      >
                        {state.label}
                      </Badge>
                    </div>
                  </div>
                </button>

                {onDelete && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className="absolute right-2 top-1/2 hidden -translate-y-1/2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive group-hover:inline-flex"
                    onClick={() => onDelete(platform.id)}
                    aria-label={`删除平台 ${platform.name}`}
                  >
                    <IconTrash className="h-4 w-4" />
                  </Button>
                )}
              </div>
            );
          })}
        </nav>
      </ScrollArea>

      <div className="border-t p-3">
        <Button variant="outline" className="h-10 w-full text-base font-normal" onClick={onAdd}>
          <IconPlus className="h-4 w-4" />
          添加
        </Button>
      </div>
    </aside>
  );
}
