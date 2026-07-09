import { IconCheck, IconChevronRight, IconSearch } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";

import type { AIModel, AIPlatform } from "@/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { getModelLogo, getProviderLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

import type { SceneDefinition } from "../agent-definitions";

interface ModelPickerDialogProps {
  scene: SceneDefinition;
  models: AIModel[];
  platforms: AIPlatform[];
  selectedModelId?: string;
  disabled: boolean;
  onSelect: (modelId: string) => void;
}

export function ModelPickerDialog({
  scene,
  models,
  platforms,
  selectedModelId,
  disabled,
  onSelect,
}: ModelPickerDialogProps) {
  const [open, setOpen] = useState(false);
  const [selectedPlatformId, setSelectedPlatformId] = useState("");
  const [keyword, setKeyword] = useState("");
  const selectedModel = models.find((model) => model.id === selectedModelId);
  const platformsWithModels = useMemo(() => {
    return platforms
      .map((platform) => ({
        platform,
        models: models.filter((model) => model.platform_id === platform.id),
      }))
      .filter((item) => item.models.length > 0);
  }, [models, platforms]);

  useEffect(() => {
    if (!open) return;
    setSelectedPlatformId(selectedModel?.platform_id || platformsWithModels[0]?.platform.id || "");
    setKeyword("");
  }, [open, platformsWithModels, selectedModel?.platform_id]);

  const selectedPlatformModels = platformsWithModels.find(
    (item) => item.platform.id === selectedPlatformId,
  )?.models;
  const visibleModels = (selectedPlatformModels ?? []).filter((model) => {
    const query = keyword.trim().toLowerCase();
    if (!query) return true;
    return `${model.display_name} ${model.model_name}`.toLowerCase().includes(query);
  });

  const handleSelect = (modelId: string) => {
    onSelect(modelId);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        选择模型
      </Button>

      <DialogContent className="max-w-5xl p-0">
        <DialogHeader className="border-b px-6 py-4">
          <DialogTitle>选择{scene.name}</DialogTitle>
          <DialogDescription>先选择模型厂商，再从该厂商已启用的模型中选择。</DialogDescription>
        </DialogHeader>

        <div className="grid min-h-[420px] grid-cols-[280px_minmax(0,1fr)] overflow-hidden">
          <aside className="flex min-h-0 flex-col bg-muted/20">
            <ScrollArea className="min-h-0 flex-1">
              <nav className="flex flex-col gap-1 p-3">
                {platformsWithModels.map(({ platform, models: platformModels }) => {
                  const selected = selectedPlatformId === platform.id;
                  const logo = getProviderLogo(platform);
                  return (
                    <button
                      key={platform.id}
                      type="button"
                      className={cn(
                        "flex items-center gap-3 rounded-md px-3 py-2.5 text-left outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring/50",
                        selected
                          ? "bg-background shadow-xs ring-1 ring-border"
                          : "hover:bg-background/70",
                      )}
                      onClick={() => setSelectedPlatformId(platform.id)}
                    >
                      <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-background ring-1 ring-border">
                        {logo ? (
                          <img src={logo} alt="" className="size-6 rounded-sm object-contain" />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {platform.name.slice(0, 2).toUpperCase()}
                          </span>
                        )}
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium">{platform.name}</span>
                        <span className="block text-xs text-muted-foreground">
                          {platformModels.length} 个可选模型
                        </span>
                      </span>
                      <IconChevronRight className="text-muted-foreground" />
                    </button>
                  );
                })}
              </nav>
            </ScrollArea>
          </aside>

          <section className="flex min-h-0 flex-col">
            <div className="p-4">
              <div className="relative">
                <IconSearch className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  placeholder="搜索模型"
                  className="pl-9"
                />
              </div>
            </div>
            <ScrollArea className="min-h-0 flex-1">
              <div className="flex flex-col gap-2 p-4">
                {visibleModels.map((model) => {
                  const selected = selectedModelId === model.id;
                  const logo = getModelLogo(`${model.display_name} ${model.model_name}`);
                  return (
                    <button
                      key={model.id}
                      type="button"
                      className={cn(
                        "flex min-h-16 items-center gap-3 rounded-md border px-4 py-3 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
                        selected ? "border-primary bg-primary/5" : "hover:bg-muted/50",
                      )}
                      onClick={() => handleSelect(model.id)}
                    >
                      <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-muted/40 ring-1 ring-border">
                        {logo ? (
                          <img src={logo} alt="" className="size-6 rounded-sm object-contain" />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {model.model_name.slice(0, 2).toUpperCase()}
                          </span>
                        )}
                      </span>
                      <span className="flex min-w-0 flex-1 flex-col">
                        <span className="truncate text-sm font-medium">
                          {model.display_name || model.model_name}
                        </span>
                        <span className="truncate font-mono text-xs text-muted-foreground">
                          {model.model_name}
                        </span>
                      </span>
                      {selected ? <IconCheck className="text-primary" /> : null}
                    </button>
                  );
                })}

                {visibleModels.length === 0 ? (
                  <div className="flex h-40 items-center justify-center rounded-md border border-dashed text-sm text-muted-foreground">
                    没有匹配的模型
                  </div>
                ) : null}
              </div>
            </ScrollArea>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}
