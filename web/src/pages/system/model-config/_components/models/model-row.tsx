import { IconTrash } from "@tabler/icons-react";
import { useState } from "react";

import { useDeleteAIModel, useUpdateAIModel } from "@/api";
import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { getModelLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

export type CandidateModel = {
  id: string;
  name: string;
  enabled: boolean;
  dbId?: string;
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
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const statusSwitchDisabled = updateModelMutation.isPending || !model.dbId;
  const deleteDisabled = !model.dbId;

  const toggleStatus = () => {
    if (!model.dbId) return;
    updateModelMutation.mutate({
      modelId: model.dbId,
      data: { status: model.enabled ? "inactive" : "active" },
    });
  };

  const handleDelete = async () => {
    if (!model.dbId) return;
    await deleteModelMutation.mutateAsync(model.dbId);
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
          </div>
          <p className="truncate text-xs text-muted-foreground">{model.id}</p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
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
            <TooltipContent>删除模型</TooltipContent>
          </Tooltip>
        </div>

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
