import type { AIModel, AIPlatform } from "@/api";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { getModelLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

import type { SceneDefinition } from "../agent-definitions";
import { ModelPickerDialog } from "./model-picker-dialog";

interface SceneRowProps {
  agentKey: string;
  scene: SceneDefinition;
  configModelId?: string;
  enabled: boolean;
  models: AIModel[];
  platforms: AIPlatform[];
  saving: boolean;
  onModelChange: (agentKey: string, scene: SceneDefinition, modelId: string) => void;
  onEnabledChange: (agentKey: string, scene: SceneDefinition, enabled: boolean) => void;
}

export function SceneRow({
  agentKey,
  scene,
  configModelId,
  enabled,
  models,
  platforms,
  saving,
  onModelChange,
  onEnabledChange,
}: SceneRowProps) {
  const hasConfig = Boolean(configModelId);
  const selectedModel = models.find((model) => model.id === configModelId);
  const selectedPlatform = selectedModel
    ? platforms.find((platform) => platform.id === selectedModel.platform_id)
    : undefined;
  const modelLogo = selectedModel
    ? getModelLogo(`${selectedModel.display_name} ${selectedModel.model_name}`)
    : undefined;
  const active = hasConfig && enabled;

  return (
    <section
      className={cn(
        "grid min-h-24 grid-cols-[minmax(220px,1fr)_minmax(260px,420px)_104px] items-center gap-5 rounded-md border bg-background px-4 py-4 transition-colors",
        active ? "hover:bg-muted/20" : "bg-muted/20",
      )}
    >
      <div className="flex min-w-0 items-start gap-3">
        <div
          className={cn(
            "mt-1 size-2.5 shrink-0 rounded-full ring-4 ring-background",
            active ? "bg-primary" : hasConfig ? "bg-muted-foreground/50" : "bg-border",
          )}
        />
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="truncate text-sm font-medium text-foreground">{scene.name}</h3>
            <Badge variant={scene.required ? "default" : "secondary"}>
              {scene.required ? "必需" : "可选"}
            </Badge>
          </div>
          <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
            {agentKey}.{scene.key}
          </div>
          <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-muted-foreground">
            {scene.description}
          </p>
        </div>
      </div>

      <div className="flex min-w-0 items-center gap-3">
        <div className="flex size-10 shrink-0 items-center justify-center rounded-md border bg-muted/30">
          {modelLogo ? (
            <img src={modelLogo} alt="" className="size-6 rounded-sm object-contain" />
          ) : (
            <span className="text-[10px] font-bold text-muted-foreground">
              {(selectedModel?.model_name || scene.name).slice(0, 2).toUpperCase()}
            </span>
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div
            className={cn(
              "truncate text-sm font-medium",
              !selectedModel && "text-muted-foreground",
            )}
          >
            {selectedModel?.display_name || selectedModel?.model_name || "未选择模型"}
          </div>
          <div className="truncate text-xs text-muted-foreground">
            {selectedModel
              ? `${selectedPlatform?.name ?? "未知厂商"} / ${selectedModel.model_name}`
              : models.length === 0
                ? "当前没有可用模型"
                : "从厂商目录中选择模型"}
          </div>
        </div>
        <ModelPickerDialog
          scene={scene}
          models={models}
          platforms={platforms}
          selectedModelId={configModelId}
          disabled={saving || models.length === 0}
          onSelect={(modelId) => onModelChange(agentKey, scene, modelId)}
        />
      </div>

      <div className="flex items-center justify-end gap-2">
        <span
          className={cn(
            "text-xs",
            hasConfig && enabled ? "text-foreground" : "text-muted-foreground",
          )}
        >
          {hasConfig && enabled ? "启用" : "停用"}
        </span>
        <Switch
          checked={hasConfig && enabled}
          disabled={saving || !hasConfig}
          onCheckedChange={(checked) => onEnabledChange(agentKey, scene, checked)}
        />
      </div>
    </section>
  );
}
