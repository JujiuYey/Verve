import { IconPlus, IconRefresh, IconSearch } from "@tabler/icons-react";
import { useMemo, useState } from "react";

import type { AIModel, ModelType } from "@/api/system/model-config";
import { useSyncModels } from "@/api/system/model-config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

import { AddModelDialog } from "./add-model-dialog";
import { ModelRow, type CandidateModel } from "./model-row";

interface ModelListProps {
  platform: {
    id: string;
    name: string;
    model_list_path: string;
    default_base_url: string;
  };
  enabledModels: AIModel[];
  initials: string;
  accent: string;
}

export function ModelList({ platform, enabledModels, initials, accent }: ModelListProps) {
  const syncModelsMutation = useSyncModels();
  const [searchText, setSearchText] = useState("");
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const models = useMemo(() => {
    return enabledModels.map(
      (model): CandidateModel => ({
        id: model.model_name,
        name: model.display_name || model.model_name,
        type: model.model_type as ModelType,
        enabled: model.status === "active",
        capabilities: model.capabilities ?? [],
        source: "enabled",
        dbId: model.id,
      }),
    );
  }, [enabledModels]);

  const keyword = searchText.trim().toLowerCase();
  const filteredModels = useMemo(() => {
    if (!keyword) return models;
    return models.filter((model) => `${model.name} ${model.id}`.toLowerCase().includes(keyword));
  }, [models, keyword]);

  const handleSync = () => {
    syncModelsMutation.mutate(platform.id);
  };

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Label className="text-base font-semibold text-foreground">已启用模型</Label>
          <span className="inline-flex min-w-8 items-center justify-center rounded-md bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
            {models.length}
          </span>
          <div className="relative">
            <IconSearch className="pointer-events-none absolute top-1/2 left-2.5 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="enabled-model-search"
              name="enabled-model-search"
              type="search"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              placeholder="搜索模型"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              aria-label="搜索模型"
              className="h-8 w-40 pl-8"
            />
          </div>
        </div>

        <div className="flex items-center gap-0">
          <Button
            type="button"
            variant="outline"
            className="h-9 rounded-r-none"
            onClick={() => void handleSync()}
            disabled={syncModelsMutation.isPending}
          >
            <IconRefresh
              className={cn("h-4 w-4", syncModelsMutation.isPending && "animate-spin")}
            />
            同步模型
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 rounded-l-none border-l-0"
            onClick={() => setAddDialogOpen(true)}
          >
            <IconPlus className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {models.length === 0 ? (
        <div className="rounded-lg border border-dashed bg-muted/10 p-8 text-center text-sm text-muted-foreground">
          暂无模型
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border bg-card">
          <div className="divide-y">
            {filteredModels.map((model) => (
              <ModelRow key={model.id} model={model} initials={initials} accent={accent} />
            ))}
          </div>
        </div>
      )}

      <AddModelDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        platformId={platform.id}
        existingModelNames={enabledModels.map((m) => m.model_name)}
      />
    </section>
  );
}
