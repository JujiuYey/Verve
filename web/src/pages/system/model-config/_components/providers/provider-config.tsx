import { IconExternalLink, IconEye, IconEyeOff, IconTrash } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import type { AIPlatform } from "@/api";
import { useUpdateAIPlatformConfig } from "@/api";
import { ConfirmDialog } from "@/components/sag-ui/confirm-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface ProviderConfigProps {
  platform: AIPlatform;
}

function getPreviewUrl(baseUrl: string, listPath: string) {
  const cleanBase = baseUrl.replace(/\/$/, "");
  const cleanPath = listPath.startsWith("/") ? listPath : `/${listPath}`;
  if (!cleanBase) return "";
  return `${cleanBase}${cleanPath}`;
}

function getConfiguredBaseUrl(platform: AIPlatform) {
  return platform.base_url || platform.default_base_url;
}

export function ProviderConfig({ platform }: ProviderConfigProps) {
  const updatePlatformMutation = useUpdateAIPlatformConfig();
  const [showApiKey, setShowApiKey] = useState(false);
  const [apiKey, setApiKey] = useState("");
  const [baseUrl, setBaseUrl] = useState(getConfiguredBaseUrl(platform));
  const [clearDialogOpen, setClearDialogOpen] = useState(false);

  useEffect(() => {
    setBaseUrl(getConfiguredBaseUrl(platform));
    setApiKey("");
    setShowApiKey(false);
    setClearDialogOpen(false);
  }, [platform]);

  const configured = Boolean(platform.api_key_hint?.trim());

  const handleSave = async () => {
    const key = apiKey.trim();
    const nextBaseUrl = baseUrl.trim();
    if (!nextBaseUrl) {
      toast.error("请填写 API 地址");
      return false;
    }
    if (!configured && !key) {
      toast.error("请先填写 API 密钥");
      return false;
    }

    await updatePlatformMutation.mutateAsync({
      platformId: platform.id,
      data: {
        base_url: nextBaseUrl,
        ...(key ? { api_key: key } : {}),
      },
    });
    setApiKey("");
    return true;
  };

  const handleResetBaseUrl = () => {
    setBaseUrl(platform.default_base_url);
  };

  const handleClearApiKey = async () => {
    await updatePlatformMutation.mutateAsync({
      platformId: platform.id,
      data: {
        base_url: baseUrl.trim() || platform.default_base_url,
        clear_api_key: true,
      },
    });
    setApiKey("");
  };

  return (
    <>
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <Label className="text-base font-semibold text-foreground">平台密钥</Label>
          {platform.api_key_url && (
            <Button variant="ghost" size="sm" className="text-muted-foreground" asChild>
              <a href={platform.api_key_url} target="_blank" rel="noreferrer">
                <IconExternalLink className="h-4 w-4" />
                获取密钥
              </a>
            </Button>
          )}
        </div>

        {platform.api_key_hint && (
          <div className="flex min-h-10 items-center gap-2 rounded-md border border-emerald-200 bg-emerald-50 px-3 text-sm text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
            <span className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-500 text-white">
              <svg
                className="h-3 w-3"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={3}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            </span>
            <span>已保存密钥：{platform.api_key_hint}</span>
          </div>
        )}

        <div className="flex gap-0">
          <div className="relative flex-1">
            <Input
              id={`api-key-${platform.id}`}
              name={`api-key-${platform.id}`}
              type={showApiKey ? "text" : "password"}
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder={platform.api_key_hint ? "留空则继续使用已保存密钥" : "请输入 API Key"}
              autoComplete="new-password"
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              data-lpignore="true"
              data-1p-ignore="true"
              className="h-11 rounded-r-none pr-10 text-base"
            />
            <button
              type="button"
              className="absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground transition hover:text-foreground"
              onClick={() => setShowApiKey((v) => !v)}
            >
              {showApiKey ? <IconEyeOff className="h-4 w-4" /> : <IconEye className="h-4 w-4" />}
            </button>
          </div>
          <Button
            type="button"
            variant="outline"
            className="h-11 rounded-l-none border-l-0 px-5"
            onClick={() => void handleSave()}
            disabled={updatePlatformMutation.isPending}
          >
            保存
          </Button>
        </div>
        {configured && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => setClearDialogOpen(true)}
            disabled={updatePlatformMutation.isPending}
          >
            <IconTrash className="h-4 w-4" />
            清空密钥
          </Button>
        )}
      </section>

      <section className="space-y-3">
        <Label className="text-base font-semibold text-foreground">API 地址</Label>
        <div className="flex gap-0">
          <Input
            id={`api-base-url-${platform.id}`}
            name={`api-base-url-${platform.id}`}
            type="url"
            value={baseUrl}
            onChange={(e) => setBaseUrl(e.target.value)}
            placeholder="https://api.example.com/v1"
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="none"
            spellCheck={false}
            className="h-11 rounded-r-none text-base"
          />
          <Button
            type="button"
            variant="outline"
            className="h-11 rounded-l-none border-l-0 px-5 text-destructive hover:text-destructive"
            disabled={baseUrl === platform.default_base_url}
            onClick={handleResetBaseUrl}
          >
            重置
          </Button>
        </div>
        <p className="text-sm text-muted-foreground">
          预览：{getPreviewUrl(baseUrl, platform.model_list_path) || "未配置"}
        </p>
      </section>

      <ConfirmDialog
        open={clearDialogOpen}
        title="清空平台密钥"
        description="清空后，该平台将无法继续使用当前密钥，已启用模型的同步和调用会受影响。"
        confirmText="确认清空"
        destructive
        onOpenChange={setClearDialogOpen}
        onConfirm={handleClearApiKey}
      />
    </>
  );
}
