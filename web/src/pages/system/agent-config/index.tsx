import { IconSettingsAutomation } from "@tabler/icons-react";
import { useMemo, useState } from "react";

import {
  useAgentModelConfigs,
  useAIModels,
  useModelPlatforms,
  useUpsertAgentModelConfig,
} from "@/api";
import { SagPage } from "@/components/sag-ui";
import { Spinner } from "@/components/ui/spinner";

import { AgentDetail } from "./_components/agent-detail";
import { AgentSidebar } from "./_components/agent-sidebar";
import { AGENTS, getSceneKey, type SceneDefinition } from "./agent-definitions";

export function AgentConfigPage() {
  const modelsQuery = useAIModels();
  const platformsQuery = useModelPlatforms();
  const configsQuery = useAgentModelConfigs();
  const upsertConfig = useUpsertAgentModelConfig();
  const [selectedAgentKey, setSelectedAgentKey] = useState(AGENTS[0]?.key ?? "");
  const models = useMemo(() => modelsQuery.data ?? [], [modelsQuery.data]);
  const platforms = useMemo(() => platformsQuery.data ?? [], [platformsQuery.data]);
  const configs = useMemo(() => configsQuery.data ?? [], [configsQuery.data]);
  const loading = modelsQuery.isLoading || platformsQuery.isLoading || configsQuery.isLoading;

  const configsByScene = useMemo(() => {
    return new Map(
      configs.map((config) => [getSceneKey(config.agent_key, config.scene_key), config]),
    );
  }, [configs]);

  const activeModels = useMemo(() => {
    return models.filter((model) => model.status === "active");
  }, [models]);

  const selectedAgent = AGENTS.find((agent) => agent.key === selectedAgentKey) ?? AGENTS[0];

  const handleModelChange = (agentKey: string, scene: SceneDefinition, modelId: string) => {
    const current = configsByScene.get(getSceneKey(agentKey, scene.key));
    upsertConfig.mutate({
      agentKey,
      sceneKey: scene.key,
      data: {
        model_id: modelId,
        params: current?.params ?? {},
        enabled: current?.enabled ?? true,
      },
    });
  };

  const handleEnabledChange = (agentKey: string, scene: SceneDefinition, enabled: boolean) => {
    const current = configsByScene.get(getSceneKey(agentKey, scene.key));
    if (!current?.model_id) return;
    upsertConfig.mutate({
      agentKey,
      sceneKey: scene.key,
      data: {
        model_id: current.model_id,
        params: current.params ?? {},
        enabled,
      },
    });
  };

  return (
    <SagPage
      title="Agent 配置"
      description="配置每个 Agent 在具体场景下使用的模型。"
      icon={<IconSettingsAutomation />}
      bodyClassName="overflow-auto"
    >
      {loading ? (
        <div className="flex h-full items-center justify-center">
          <Spinner />
        </div>
      ) : (
        <div className="grid h-full min-h-[640px] grid-cols-[320px_minmax(0,1fr)] overflow-hidden rounded-lg border bg-background">
          <AgentSidebar
            selectedAgent={selectedAgent}
            configsByScene={configsByScene}
            onSelectAgent={setSelectedAgentKey}
          />

          {selectedAgent ? (
            <AgentDetail
              agent={selectedAgent}
              activeModels={activeModels}
              platforms={platforms}
              configsByScene={configsByScene}
              saving={upsertConfig.isPending}
              onModelChange={handleModelChange}
              onEnabledChange={handleEnabledChange}
            />
          ) : null}
        </div>
      )}
    </SagPage>
  );
}
