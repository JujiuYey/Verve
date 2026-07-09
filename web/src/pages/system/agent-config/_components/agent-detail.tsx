import { IconBrain } from "@tabler/icons-react";

import type { AIModel, AIPlatform } from "@/api";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";

import {
  getAgentStatus,
  getSceneKey,
  type AgentDefinition,
  type SceneConfigState,
  type SceneDefinition,
} from "../agent-definitions";
import { SceneRow } from "./scene-row";
import { StatusBadge } from "./status-badge";

interface AgentDetailProps {
  agent: AgentDefinition;
  activeModels: AIModel[];
  platforms: AIPlatform[];
  configsByScene: Map<string, SceneConfigState & { params?: Record<string, unknown> }>;
  saving: boolean;
  onModelChange: (agentKey: string, scene: SceneDefinition, modelId: string) => void;
  onEnabledChange: (agentKey: string, scene: SceneDefinition, enabled: boolean) => void;
}

export function AgentDetail({
  agent,
  activeModels,
  platforms,
  configsByScene,
  saving,
  onModelChange,
  onEnabledChange,
}: AgentDetailProps) {
  const Icon = agent.icon;
  const status = getAgentStatus(agent, configsByScene);
  const missingRequiredScenes = agent.scenes.filter((scene) => {
    if (!scene.required) return false;
    const config = configsByScene.get(getSceneKey(agent.key, scene.key));
    return !config?.model_id || !config.enabled;
  });

  return (
    <main className="flex min-h-0 flex-col">
      <header className="flex shrink-0 items-start justify-between gap-4 border-b px-6 py-5">
        <div className="flex min-w-0 items-start gap-4">
          <div className="flex size-11 shrink-0 items-center justify-center rounded-md border bg-muted/40">
            <Icon />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="truncate text-xl font-semibold tracking-tight">{agent.name}</h2>
              <StatusBadge status={status} />
              <Badge variant="outline">{agent.runtime}</Badge>
            </div>
            <p className="mt-1 max-w-3xl text-sm text-muted-foreground">{agent.description}</p>
          </div>
        </div>
      </header>

      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-auto p-6">
        {missingRequiredScenes.length > 0 ? (
          <Alert>
            <IconBrain />
            <AlertTitle>必需模型未配置</AlertTitle>
            <AlertDescription>
              {missingRequiredScenes.map((scene) => scene.name).join("、")} 需要先绑定模型。
            </AlertDescription>
          </Alert>
        ) : null}

        <div className="flex flex-col gap-3">
          {agent.scenes.map((scene) => {
            const config = configsByScene.get(getSceneKey(agent.key, scene.key));
            return (
              <SceneRow
                key={scene.key}
                agentKey={agent.key}
                scene={scene}
                configModelId={config?.model_id}
                enabled={config?.enabled ?? true}
                models={activeModels}
                platforms={platforms}
                saving={saving}
                onModelChange={onModelChange}
                onEnabledChange={onEnabledChange}
              />
            );
          })}
        </div>
      </div>
    </main>
  );
}
