import { useState } from "react";

import type { CreateAIPlatformRequest } from "@/api/ai/model-config";
import { useCreateAIPlatform } from "@/api/ai/model-config";
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

const PROVIDER_TYPES = [
  { value: "openai_compatible", label: "OpenAI 兼容" },
  { value: "custom", label: "自定义" },
] as const;

export function CreatePlatformDialog({ open, onOpenChange, onCreated }: CreatePlatformDialogProps) {
  const createPlatformMutation = useCreateAIPlatform();
  const [name, setName] = useState("");
  const [providerType, setProviderType] = useState<string>("openai_compatible");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [docsUrl, setDocsUrl] = useState("");

  const reset = () => {
    setName("");
    setProviderType("openai_compatible");
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
      provider_type: providerType,
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
          <DialogDescription>添加一个自定义模型平台，支持 OpenAI 兼容接口。</DialogDescription>
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
            <Label>接口类型</Label>
            <div className="grid grid-cols-2 gap-2">
              {PROVIDER_TYPES.map((type) => (
                <button
                  key={type.value}
                  type="button"
                  onClick={() => setProviderType(type.value)}
                  className={`rounded-md border py-2 text-sm font-medium transition-colors ${
                    providerType === type.value
                      ? "border-primary bg-primary/10 text-primary"
                      : "border-border text-muted-foreground hover:border-primary/50"
                  }`}
                >
                  {type.label}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label>API 地址</Label>
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
