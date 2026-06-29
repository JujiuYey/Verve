import { useState } from "react";

import type { CreateAIPlatformRequest } from "@/api/system/model-config";
import { useCreateAIPlatform } from "@/api/system/model-config";
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

interface CreatePlatformDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: (platformId: string) => void;
}

export function CreatePlatformDialog({ open, onOpenChange, onCreated }: CreatePlatformDialogProps) {
  const createPlatformMutation = useCreateAIPlatform();
  const [name, setName] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [docsUrl, setDocsUrl] = useState("");

  const reset = () => {
    setName("");
    setBaseUrl("");
    setApiKey("");
    setDocsUrl("");
  };

  const handleOpenChange = (value: boolean) => {
    if (!value) reset();
    onOpenChange(value);
  };

  const handleSubmit = async () => {
    const trimmedName = name.trim();
    const trimmedBaseUrl = baseUrl.trim();
    const trimmedApiKey = apiKey.trim();
    if (!trimmedName || !trimmedBaseUrl || !trimmedApiKey) return;

    const payload: CreateAIPlatformRequest = {
      name: trimmedName,
      provider_type: "openai_compatible",
      default_base_url: trimmedBaseUrl,
      base_url: trimmedBaseUrl,
      api_key: trimmedApiKey,
      model_list_path: "/models",
      auth_scheme: "bearer",
      ...(docsUrl.trim() ? { docs_url: docsUrl.trim() } : {}),
    };

    const platform = await createPlatformMutation.mutateAsync(payload);
    reset();
    onOpenChange(false);
    onCreated?.(platform.id);
  };

  const canSubmit =
    name.trim() && baseUrl.trim() && apiKey.trim() && !createPlatformMutation.isPending;
  const showApiKeyRequired = !apiKey.trim();

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>新增平台</DialogTitle>
          <DialogDescription>添加 OpenAI 兼容模型平台，用于 Eino 模型调用。</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>平台名称</Label>
            <Input
              id="new-model-platform-name"
              name="new-model-platform-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如: My LLM Platform"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              className="h-10"
            />
          </div>

          <div className="space-y-2">
            <Label>OpenAI 兼容接口地址</Label>
            <Input
              id="new-model-platform-base-url"
              name="new-model-platform-base-url"
              type="url"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="https://api.example.com/v1"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              className="h-10"
            />
          </div>

          <div className="space-y-2">
            <Label>平台密钥</Label>
            <Input
              id="new-model-platform-api-key"
              name="new-model-platform-api-key"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="请输入 API Key"
              autoComplete="new-password"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              aria-invalid={showApiKeyRequired}
              className="h-10"
            />
            {showApiKeyRequired && <p className="text-sm text-destructive">API Key 为必填项</p>}
          </div>

          <div className="space-y-2">
            <Label>
              文档地址
              <span className="ml-1 text-xs text-muted-foreground">（可选）</span>
            </Label>
            <Input
              id="new-model-platform-docs-url"
              name="new-model-platform-docs-url"
              type="url"
              value={docsUrl}
              onChange={(e) => setDocsUrl(e.target.value)}
              placeholder="https://docs.example.com"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              className="h-10"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            取消
          </Button>
          <Button onClick={() => void handleSubmit()} disabled={!canSubmit}>
            创建
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
