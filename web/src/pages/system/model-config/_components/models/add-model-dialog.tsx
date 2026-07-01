import { useState } from "react";
import { toast } from "sonner";

import type { ModelType } from "@/api";
import { useCreateAIModel } from "@/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface AddModelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  platformId: string;
  existingModelNames: string[];
}

export function AddModelDialog({
  open,
  onOpenChange,
  platformId,
  existingModelNames,
}: AddModelDialogProps) {
  const createModelMutation = useCreateAIModel();
  const [modelId, setModelId] = useState("");
  const [modelType, setModelType] = useState<ModelType>("chat");

  const handleAdd = async () => {
    const id = modelId.trim();
    if (!id) return;
    if (existingModelNames.includes(id)) {
      toast.error("模型已启用");
      return;
    }

    await createModelMutation.mutateAsync({
      platform_id: platformId,
      model_name: id,
      display_name: id,
      model_type: modelType,
      capabilities:
        modelType === "embedding" ? ["embedding"] : modelType === "rerank" ? ["rerank"] : [],
      source: "manual",
    });
    setModelId("");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>新增模型</DialogTitle>
          <DialogDescription>填写模型 ID 并启用到当前模型平台。</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>模型 ID</Label>
            <Input
              id="new-enabled-model-id"
              name="new-enabled-model-id"
              value={modelId}
              onChange={(e) => setModelId(e.target.value)}
              placeholder="例如: qwen-plus"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              className="h-10"
            />
          </div>

          <div className="space-y-2">
            <Label>类型</Label>
            <div className="grid grid-cols-3 gap-2">
              {(["chat", "embedding", "rerank"] as const).map((type) => (
                <button
                  key={type}
                  type="button"
                  onClick={() => setModelType(type)}
                  className={cn(
                    "rounded-md border py-2 text-sm font-medium transition-colors",
                    modelType === type
                      ? "border-primary bg-primary/10 text-primary"
                      : "border-border text-muted-foreground hover:border-primary/50",
                  )}
                >
                  {type === "chat" ? "对话" : type === "embedding" ? "向量" : "重排"}
                </button>
              ))}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={() => void handleAdd()}
            disabled={!modelId.trim() || createModelMutation.isPending}
          >
            添加
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
