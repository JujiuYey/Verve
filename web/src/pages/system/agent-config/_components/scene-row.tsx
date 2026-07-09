import type { AIModel, AIPlatform } from "@/api";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
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

  return (
    <div className="grid min-h-20 grid-cols-[minmax(220px,1fr)_minmax(280px,440px)_96px] items-center gap-4 px-4 py-3">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <div className="truncate text-sm font-medium text-foreground">{scene.name}</div>
          <Badge variant={scene.required ? "default" : "secondary"}>
            {scene.required ? "必需" : "可选"}
          </Badge>
        </div>
        <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
          {agentKey}.{scene.key}
        </div>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{scene.description}</p>
      </div>

      <div className="flex min-w-0 items-center justify-between gap-3 rounded-md border bg-background px-3 py-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
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
    </div>
  );
}
