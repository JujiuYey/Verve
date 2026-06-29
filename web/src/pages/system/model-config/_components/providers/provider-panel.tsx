import { IconExternalLink } from "@tabler/icons-react";

import type { AIModel, AIPlatform } from "@/api/system/model-config";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { getPlatformAccent } from "@/constants/model-config";
import { getProviderLogo } from "@/lib/model-logos";

import { ModelList } from "../models/model-list";
import { ProviderConfig } from "./provider-config";

interface ProviderPanelProps {
  platform: AIPlatform;
  enabledModels: AIModel[];
}

function getInitials(name: string) {
  return name.slice(0, 2).toUpperCase();
}

function getPlatformState(configured: boolean, modelCount: number) {
  if (!configured) {
    return {
      label: "未配置 API Key",
      badge: "未接入",
      badgeVariant: "secondary" as const,
    };
  }
  if (modelCount === 0) {
    return {
      label: "已保存配置，未启用模型",
      badge: "已接入",
      badgeVariant: "outline" as const,
    };
  }
  return {
    label: "已启用模型，可供业务调用",
    badge: "可用",
    badgeVariant: "default" as const,
  };
}

export function ProviderPanel({ platform, enabledModels }: ProviderPanelProps) {
  const initials = getInitials(platform.name);
  const accent = getPlatformAccent(platform.id);
  const providerLogo = getProviderLogo(platform);

  const configured = Boolean(platform.api_key_hint?.trim());
  const activeModelCount = enabledModels.filter((model) => model.status === "active").length;
  const platformState = getPlatformState(configured, activeModelCount);

  return (
    <main className="flex min-h-0 flex-1 flex-col bg-background">
      <header className="flex items-center justify-between gap-4 px-8 py-4">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-background shadow-xs ring-1 ring-border">
            {providerLogo ? (
              <img
                src={providerLogo}
                alt=""
                className="h-7 w-7 rounded-sm object-contain"
                draggable={false}
              />
            ) : (
              <span className="text-sm font-bold text-muted-foreground">{initials}</span>
            )}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h2 className="truncate text-xl font-semibold tracking-tight text-foreground">
                {platform.name}
              </h2>
              {platform.docs_url && (
                <Button variant="ghost" size="icon-sm" className="text-muted-foreground" asChild>
                  <a href={platform.docs_url} target="_blank" rel="noreferrer">
                    <IconExternalLink className="h-4 w-4" />
                  </a>
                </Button>
              )}
            </div>
            <p className="truncate text-sm text-muted-foreground">{platformState.label}</p>
          </div>
        </div>

        <Badge variant={platformState.badgeVariant}>{platformState.badge}</Badge>
      </header>

      <ScrollArea className="min-h-0 flex-1">
        <div className="space-y-8 px-8 py-2">
          <ProviderConfig platform={platform} />
          <ModelList
            platform={platform}
            enabledModels={enabledModels}
            initials={initials}
            accent={accent}
          />
        </div>
      </ScrollArea>
    </main>
  );
}
