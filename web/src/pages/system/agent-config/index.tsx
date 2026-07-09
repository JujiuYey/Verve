import { IconSettingsAutomation } from "@tabler/icons-react";
import { useMemo } from "react";

import type { AIModel, ModelType } from "@/api";
import { useAgentModelConfigs, useAIModels, useUpsertAgentModelConfig } from "@/api";
import { SagPage } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";

type SceneDefinition = {
  key: string;
  name: string;
  type: ModelType;
  required?: boolean;
};

type AgentDefinition = {
  key: string;
  name: string;
  description: string;
  scenes: SceneDefinition[];
};

const AGENTS: AgentDefinition[] = [
  {
    key: "wiki_rag",
    name: "Wiki RAG",
    description: "知识库解析、检索与问答生成链路。",
    scenes: [
      { key: "embedding", name: "向量化", type: "embedding", required: true },
      { key: "rerank", name: "重排", type: "rerank" },
      { key: "answer", name: "回答生成", type: "chat" },
    ],
  },
  {
    key: "coach",
    name: "Coach",
    description: "学习陪练对话、工具调用与结果总结。",
    scenes: [
      { key: "chat", name: "对话", type: "chat" },
      { key: "tool_call", name: "工具调用", type: "chat" },
      { key: "summary", name: "总结", type: "chat" },
    ],
  },
];

const modelTypeLabel: Record<ModelType, string> = {
  chat: "对话",
  embedding: "向量",
  rerank: "重排",
};

export function AgentConfigPage() {
  const modelsQuery = useAIModels();
  const configsQuery = useAgentModelConfigs();
  const upsertConfig = useUpsertAgentModelConfig();
  const models = useMemo(() => modelsQuery.data ?? [], [modelsQuery.data]);
  const configs = useMemo(() => configsQuery.data ?? [], [configsQuery.data]);
  const loading = modelsQuery.isLoading || configsQuery.isLoading;

  const configsByScene = useMemo(() => {
    return new Map(configs.map((config) => [`${config.agent_key}.${config.scene_key}`, config]));
  }, [configs]);

  const activeModels = useMemo(() => {
    return models.filter((model) => model.status === "active");
  }, [models]);

  const handleModelChange = (agentKey: string, scene: SceneDefinition, modelId: string) => {
    const current = configsByScene.get(`${agentKey}.${scene.key}`);
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
    const current = configsByScene.get(`${agentKey}.${scene.key}`);
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
        <div className="flex max-w-5xl flex-col gap-4">
          {AGENTS.map((agent) => (
            <Card key={agent.key} className="rounded-lg">
              <CardHeader>
                <CardTitle>{agent.name}</CardTitle>
                <CardDescription>{agent.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="divide-y rounded-md border">
                  {agent.scenes.map((scene) => {
                    const config = configsByScene.get(`${agent.key}.${scene.key}`);
                    const sceneModels = activeModels.filter(
                      (model) => model.model_type === scene.type,
                    );
                    return (
                      <SceneRow
                        key={scene.key}
                        agentKey={agent.key}
                        scene={scene}
                        configModelId={config?.model_id}
                        enabled={config?.enabled ?? true}
                        models={sceneModels}
                        saving={upsertConfig.isPending}
                        onModelChange={handleModelChange}
                        onEnabledChange={handleEnabledChange}
                      />
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </SagPage>
  );
}

interface SceneRowProps {
  agentKey: string;
  scene: SceneDefinition;
  configModelId?: string;
  enabled: boolean;
  models: AIModel[];
  saving: boolean;
  onModelChange: (agentKey: string, scene: SceneDefinition, modelId: string) => void;
  onEnabledChange: (agentKey: string, scene: SceneDefinition, enabled: boolean) => void;
}

function SceneRow({
  agentKey,
  scene,
  configModelId,
  enabled,
  models,
  saving,
  onModelChange,
  onEnabledChange,
}: SceneRowProps) {
  const hasConfig = Boolean(configModelId);

  return (
    <div className="grid min-h-16 grid-cols-[minmax(180px,1fr)_minmax(280px,420px)_96px] items-center gap-4 px-4 py-3">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <div className="truncate text-sm font-medium text-foreground">{scene.name}</div>
          <Badge variant={scene.required ? "default" : "secondary"}>
            {scene.required ? "必需" : "可选"}
          </Badge>
          <Badge variant="outline">{modelTypeLabel[scene.type]}</Badge>
        </div>
        <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
          {agentKey}.{scene.key}
        </div>
      </div>

      <Select
        value={configModelId}
        onValueChange={(value) => onModelChange(agentKey, scene, value)}
        disabled={saving || models.length === 0}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={models.length === 0 ? "暂无可用模型" : "选择模型"} />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {models.map((model) => (
              <SelectItem key={model.id} value={model.id}>
                <span className="flex min-w-0 flex-col">
                  <span className="truncate">{model.display_name || model.model_name}</span>
                  <span className="truncate font-mono text-xs text-muted-foreground">
                    {model.model_name}
                  </span>
                </span>
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>

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
