import { IconStar, IconTrash } from "@tabler/icons-react";
import { useEffect, useState } from "react";

import { useDeleteAIModel, useUpdateAIModel } from "@/api/ai/model-config";
import type { ModelCapability, ModelType } from "@/api/ai/model-config";
import type { ModelUseCase } from "@/api/ai/scene-default";
import { useSetSceneDefault } from "@/api/ai/scene-default";
import { ConfirmDialog } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { getModelLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

export type CandidateModel = {
  id: string;
  name: string;
  type: ModelType;
  enabled: boolean;
  capabilities: ModelCapability[];
  source: "enabled";
  dbId?: string;
  boundScenes?: ModelUseCase[];
};

const capabilityMeta: Record<
  ModelCapability,
  { label: string; className: string; icon: React.ComponentType<{ className?: string }> }
> = {
  vision: {
    label: "视觉",
    className: "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300",
    icon: () => null,
  },
  reasoning: {
    label: "推理",
    className: "bg-indigo-100 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300",
    icon: () => null,
  },
  tool: {
    label: "工具",
    className: "bg-orange-100 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300",
    icon: () => null,
  },
  embedding: {
    label: "向量",
    className: "bg-cyan-100 text-cyan-700 dark:bg-cyan-500/15 dark:text-cyan-300",
    icon: () => null,
  },
  rerank: {
    label: "重排",
    className: "bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300",
    icon: () => null,
  },
};

const USE_CASE_LABELS: Record<ModelUseCase, string> = {
  keyword_expand: "关键词扩展",
  content_generate: "内容生成",
};

interface ModelRowProps {
  model: CandidateModel;
  initials: string;
  accent: string;
}

export function ModelRow({ model, initials, accent }: ModelRowProps) {
  const logo = getModelLogo(`${model.id} ${model.name}`);
  const deleteModelMutation = useDeleteAIModel();
  const updateModelMutation = useUpdateAIModel();
  const setDefaultMutation = useSetSceneDefault();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedScenes, setSelectedScenes] = useState<ModelUseCase[]>(model.boundScenes ?? []);
  const hasBoundScenes = Boolean(model.boundScenes?.length);
  const sceneSettingsDisabled = !model.dbId || !model.enabled;
  const statusSwitchDisabled = updateModelMutation.isPending || !model.dbId || hasBoundScenes;
  const deleteDisabled = !model.dbId || hasBoundScenes;

  useEffect(() => {
    setSelectedScenes(model.boundScenes ?? []);
  }, [model.boundScenes]);

  const toggleStatus = () => {
    if (!model.dbId) return;
    updateModelMutation.mutate({
      modelId: model.dbId,
      data: { status: model.enabled ? "inactive" : "active" },
    });
  };

  const handleConfirm = () => {
    if (!model.dbId) return;
    setDefaultMutation.mutate({ modelId: model.dbId, scenes: selectedScenes });
    setDialogOpen(false);
  };

  const handleDelete = async () => {
    if (!model.dbId) return;
    await deleteModelMutation.mutateAsync(model.dbId);
  };

  const toggleScene = (scene: ModelUseCase) => {
    setSelectedScenes((prev) =>
      prev.includes(scene) ? prev.filter((s) => s !== scene) : [...prev, scene],
    );
  };

  return (
    <TooltipProvider>
      <div className="flex min-h-14 items-center gap-3 px-5 py-3">
        <div
          className={cn(
            "flex h-8 w-8 shrink-0 items-center justify-center rounded-md shadow-xs",
            logo ? "bg-transparent" : `bg-gradient-to-br text-white ${accent}`,
          )}
        >
          {logo ? (
            <img
              src={logo}
              alt=""
              className="h-8 w-8 rounded-sm object-contain"
              draggable={false}
            />
          ) : (
            <span className="text-[10px] font-bold">{initials}</span>
          )}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <p className="truncate text-base font-medium text-foreground">{model.name}</p>
            {model.boundScenes?.map((scene) => (
              <span
                key={scene}
                className="inline-flex items-center gap-0.5 rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-500/15 dark:text-amber-300"
              >
                <IconStar className="h-2.5 w-2.5" />
                {USE_CASE_LABELS[scene]}
              </span>
            ))}
          </div>
          <p className="truncate text-xs text-muted-foreground">{model.id}</p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          {model.capabilities?.map((capability) => (
            <CapabilityPill key={capability} capability={capability} />
          ))}
          <Badge variant={model.type === "chat" ? "secondary" : "outline"}>
            {model.type === "chat" ? "对话" : model.type === "embedding" ? "向量" : "重排"}
          </Badge>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="text-muted-foreground hover:text-amber-500"
                onClick={() => setDialogOpen(true)}
                disabled={sceneSettingsDisabled}
              >
                <IconStar className="h-4 w-4" />
                场景
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {sceneSettingsDisabled && !model.enabled
                ? "启用模型后才可设置使用场景"
                : "设置模型使用场景"}
            </TooltipContent>
          </Tooltip>
          <Switch
            checked={model.enabled}
            onCheckedChange={toggleStatus}
            disabled={statusSwitchDisabled}
          />
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                onClick={() => setDeleteDialogOpen(true)}
                disabled={deleteDisabled}
                aria-label={`删除模型 ${model.name}`}
              >
                <IconTrash className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {hasBoundScenes ? "已绑定使用场景的模型不可删除" : "删除模型"}
            </TooltipContent>
          </Tooltip>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogContent className="sm:max-w-sm">
            <DialogHeader>
              <DialogTitle>设为默认模型</DialogTitle>
              <DialogDescription>选择此模型用于哪些业务场景。</DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              {(Object.entries(USE_CASE_LABELS) as [ModelUseCase, string][]).map(
                ([value, label]) => (
                  <label
                    key={value}
                    className="flex cursor-pointer items-center gap-3 rounded-md border p-3 hover:bg-muted/50"
                  >
                    <Checkbox
                      checked={selectedScenes.includes(value)}
                      onCheckedChange={() => toggleScene(value)}
                    />
                    <span className="text-sm font-medium">{label}</span>
                  </label>
                ),
              )}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                取消
              </Button>
              <Button onClick={handleConfirm} disabled={setDefaultMutation.isPending}>
                确认
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <ConfirmDialog
          open={deleteDialogOpen}
          title="删除模型"
          description={`确定要删除模型「${model.name}」吗？删除后无法恢复。`}
          confirmText="删除"
          destructive
          onOpenChange={setDeleteDialogOpen}
          onConfirm={handleDelete}
        />
      </div>
    </TooltipProvider>
  );
}

function CapabilityPill({ capability }: { capability: ModelCapability }) {
  const meta = capabilityMeta[capability];
  if (!meta) return null;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span
          className={cn(
            "inline-flex h-7 w-9 items-center justify-center rounded-md",
            meta.className,
          )}
        >
          <span className="text-xs">{meta.label[0]}</span>
        </span>
      </TooltipTrigger>
      <TooltipContent>{meta.label}</TooltipContent>
    </Tooltip>
  );
}
