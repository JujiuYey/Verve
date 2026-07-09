import {
  IconBook2,
  IconBrain,
  IconChecklist,
  IconCheck,
  IconChevronRight,
  IconRoute,
  IconSearch,
  IconSettingsAutomation,
  IconStack2,
  IconTargetArrow,
  IconUserQuestion,
} from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";

import type { AIModel, AIPlatform, ModelType } from "@/api";
import {
  useAgentModelConfigs,
  useAIModels,
  useModelPlatforms,
  useUpsertAgentModelConfig,
} from "@/api";
import { SagPage } from "@/components/sag-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import { getProviderLogo } from "@/lib/model-logos";
import { cn } from "@/lib/utils";

type SceneDefinition = {
  key: string;
  name: string;
  type: ModelType;
  required?: boolean;
  description: string;
};

type AgentDefinition = {
  key: string;
  name: string;
  shortName: string;
  description: string;
  runtime: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  scenes: SceneDefinition[];
};

const AGENTS: AgentDefinition[] = [
  {
    key: "guide",
    name: "导学 Agent",
    shortName: "导学",
    description: "阅读资料并生成掌握目标、阅读步骤和自检问题。",
    runtime: "GuideAgent",
    icon: IconBook2,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        type: "chat",
        required: true,
        description: "用于生成导学摘要、掌握目标和练习重点。",
      },
    ],
  },
  {
    key: "objective_generator",
    name: "学习小节生成 Agent",
    shortName: "小节生成",
    description: "把 Markdown 学习资料拆成适合费曼学习的小节。",
    runtime: "ObjectiveGeneratorAgent",
    icon: IconTargetArrow,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        type: "chat",
        required: true,
        description: "用于结构化生成学习小节和练习入口。",
      },
    ],
  },
  {
    key: "coach",
    name: "学习调度 Agent",
    shortName: "调度",
    description: "查询学习上下文并决定下一步学习动作。",
    runtime: "LearningCoach",
    icon: IconRoute,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        type: "chat",
        required: true,
        description: "用于对话、工具调用决策和下一步动作规划。",
      },
    ],
  },
  {
    key: "tutor",
    name: "费曼陪练 Agent",
    shortName: "陪练",
    description: "围绕当前资料追问、提示并帮助学习者复述。",
    runtime: "TutorAgent",
    icon: IconUserQuestion,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        type: "chat",
        required: true,
        description: "用于费曼练习中的追问、反馈和引导。",
      },
    ],
  },
  {
    key: "examiner",
    name: "学习监督 Agent",
    shortName: "监督",
    description: "判断一次作答是否达标，并给出改进建议。",
    runtime: "ExaminerAgent",
    icon: IconChecklist,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        type: "chat",
        required: true,
        description: "用于评估作答质量和生成判定结果。",
      },
    ],
  },
  {
    key: "wiki_rag",
    name: "Wiki RAG",
    shortName: "知识库",
    description: "知识库解析与向量检索能力。",
    runtime: "Wiki Retrieval",
    icon: IconStack2,
    scenes: [
      {
        key: "embedding",
        name: "向量化模型",
        type: "embedding",
        required: true,
        description: "解析知识库时生成文档块向量；未配置时不会创建解析任务。",
      },
    ],
  },
];

const modelTypeLabel: Partial<Record<ModelType, string>> = {
  chat: "对话",
  embedding: "向量",
};

function getSceneKey(agentKey: string, sceneKey: string) {
  return `${agentKey}.${sceneKey}`;
}

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
  const totalRequiredScenes = AGENTS.flatMap((agent) =>
    agent.scenes.map((scene) => ({ agent, scene })),
  ).filter(({ scene }) => scene.required);
  const configuredRequiredScenes = totalRequiredScenes.filter(({ agent, scene }) => {
    const config = configsByScene.get(getSceneKey(agent.key, scene.key));
    return Boolean(config?.model_id && config.enabled);
  });
  const activeModelCountByType = activeModels.reduce<Partial<Record<ModelType, number>>>(
    (acc, model) => {
      acc[model.model_type] = (acc[model.model_type] ?? 0) + 1;
      return acc;
    },
    {},
  );

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
          <aside className="flex min-h-0 flex-col border-r bg-muted/20">
            <div className="flex flex-col gap-4 border-b p-4">
              <div className="grid grid-cols-3 gap-2">
                <Metric label="Agent" value={AGENTS.length} />
                <Metric
                  label="必需"
                  value={`${configuredRequiredScenes.length}/${totalRequiredScenes.length}`}
                />
                <Metric label="模型" value={activeModels.length} />
              </div>
              <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                <span>对话 {activeModelCountByType.chat ?? 0}</span>
                <span>向量 {activeModelCountByType.embedding ?? 0}</span>
              </div>
            </div>

            <nav className="flex min-h-0 flex-1 flex-col gap-1 overflow-auto p-3">
              {AGENTS.map((agent) => {
                const status = getAgentStatus(agent, configsByScene);
                const Icon = agent.icon;
                const selected = selectedAgent?.key === agent.key;
                return (
                  <button
                    key={agent.key}
                    type="button"
                    className={cn(
                      "flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
                      selected
                        ? "bg-background shadow-xs ring-1 ring-border"
                        : "hover:bg-background/70",
                    )}
                    onClick={() => setSelectedAgentKey(agent.key)}
                  >
                    <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-background ring-1 ring-border">
                      <Icon />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium">{agent.shortName}</span>
                      <span className="block truncate font-mono text-xs text-muted-foreground">
                        {agent.runtime}
                      </span>
                    </span>
                    <StatusBadge status={status} />
                  </button>
                );
              })}
            </nav>
          </aside>

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

function Metric({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex min-h-16 flex-col justify-center gap-1 rounded-md border bg-background px-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="text-lg font-semibold leading-none">{value}</div>
    </div>
  );
}

type AgentStatus = "ready" | "partial" | "missing";

function getAgentStatus(
  agent: AgentDefinition,
  configsByScene: Map<string, { model_id: string; enabled: boolean }>,
): AgentStatus {
  const requiredScenes = agent.scenes.filter((scene) => scene.required);
  if (requiredScenes.length === 0) {
    const configuredOptional = agent.scenes.some((scene) => {
      const config = configsByScene.get(getSceneKey(agent.key, scene.key));
      return Boolean(config?.model_id && config.enabled);
    });
    return configuredOptional ? "ready" : "partial";
  }
  const configuredRequired = requiredScenes.filter((scene) => {
    const config = configsByScene.get(getSceneKey(agent.key, scene.key));
    return Boolean(config?.model_id && config.enabled);
  });
  if (configuredRequired.length === requiredScenes.length) return "ready";
  if (configuredRequired.length > 0) return "partial";
  return "missing";
}

function StatusBadge({ status }: { status: AgentStatus }) {
  if (status === "ready") return <Badge variant="default">就绪</Badge>;
  if (status === "partial") return <Badge variant="secondary">部分</Badge>;
  return <Badge variant="outline">未配置</Badge>;
}

interface AgentDetailProps {
  agent: AgentDefinition;
  activeModels: AIModel[];
  platforms: AIPlatform[];
  configsByScene: Map<
    string,
    { model_id: string; enabled: boolean; params?: Record<string, unknown> }
  >;
  saving: boolean;
  onModelChange: (agentKey: string, scene: SceneDefinition, modelId: string) => void;
  onEnabledChange: (agentKey: string, scene: SceneDefinition, enabled: boolean) => void;
}

function AgentDetail({
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

        <Card className="rounded-lg">
          <CardHeader>
            <CardTitle>模型绑定</CardTitle>
            <CardDescription>
              每个场景只保存“使用哪个模型”，模型资产仍在模型配置里维护。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col rounded-md border">
              {agent.scenes.map((scene, index) => {
                const config = configsByScene.get(getSceneKey(agent.key, scene.key));
                const sceneModels = activeModels.filter((model) => model.model_type === scene.type);
                return (
                  <div key={scene.key}>
                    {index > 0 ? <Separator /> : null}
                    <SceneRow
                      agentKey={agent.key}
                      scene={scene}
                      configModelId={config?.model_id}
                      enabled={config?.enabled ?? true}
                      models={sceneModels}
                      platforms={platforms}
                      saving={saving}
                      onModelChange={onModelChange}
                      onEnabledChange={onEnabledChange}
                    />
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

interface SceneRowProps {
  agentKey: string;
  scene: SceneDefinition;
  configModelId?: string;
  enabled: boolean;
  models: AIModel[];
  platforms: AIPlatform[];
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
  platforms,
  saving,
  onModelChange,
  onEnabledChange,
}: SceneRowProps) {
  const hasConfig = Boolean(configModelId);
  const selectedModel = models.find((model) => model.id === configModelId);
  const selectedPlatform = selectedModel
    ? platforms.find((platform) => platform.id === selectedModel.platform_id)
    : undefined;

  return (
    <div className="grid min-h-20 grid-cols-[minmax(220px,1fr)_minmax(280px,440px)_96px] items-center gap-4 px-4 py-3">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <div className="truncate text-sm font-medium text-foreground">{scene.name}</div>
          <Badge variant={scene.required ? "default" : "secondary"}>
            {scene.required ? "必需" : "可选"}
          </Badge>
          <Badge variant="outline">{modelTypeLabel[scene.type] ?? scene.type}</Badge>
        </div>
        <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
          {agentKey}.{scene.key}
        </div>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{scene.description}</p>
      </div>

      <div className="flex min-w-0 items-center justify-between gap-3 rounded-md border bg-background px-3 py-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
            {selectedModel?.display_name || selectedModel?.model_name || "未选择模型"}
          </div>
          <div className="truncate text-xs text-muted-foreground">
            {selectedModel
              ? `${selectedPlatform?.name ?? "未知厂商"} / ${selectedModel.model_name}`
              : models.length === 0
                ? "当前没有可用模型"
                : "从厂商目录中选择模型"}
          </div>
        </div>
        <ModelPickerDialog
          scene={scene}
          models={models}
          platforms={platforms}
          selectedModelId={configModelId}
          disabled={saving || models.length === 0}
          onSelect={(modelId) => onModelChange(agentKey, scene, modelId)}
        />
      </div>

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

interface ModelPickerDialogProps {
  scene: SceneDefinition;
  models: AIModel[];
  platforms: AIPlatform[];
  selectedModelId?: string;
  disabled: boolean;
  onSelect: (modelId: string) => void;
}

function ModelPickerDialog({
  scene,
  models,
  platforms,
  selectedModelId,
  disabled,
  onSelect,
}: ModelPickerDialogProps) {
  const [open, setOpen] = useState(false);
  const [selectedPlatformId, setSelectedPlatformId] = useState("");
  const [keyword, setKeyword] = useState("");
  const selectedModel = models.find((model) => model.id === selectedModelId);
  const platformsWithModels = useMemo(() => {
    return platforms
      .map((platform) => ({
        platform,
        models: models.filter((model) => model.platform_id === platform.id),
      }))
      .filter((item) => item.models.length > 0);
  }, [models, platforms]);

  useEffect(() => {
    if (!open) return;
    setSelectedPlatformId(selectedModel?.platform_id || platformsWithModels[0]?.platform.id || "");
    setKeyword("");
  }, [open, platformsWithModels, selectedModel?.platform_id]);

  const selectedPlatformModels = platformsWithModels.find(
    (item) => item.platform.id === selectedPlatformId,
  )?.models;
  const visibleModels = (selectedPlatformModels ?? []).filter((model) => {
    const query = keyword.trim().toLowerCase();
    if (!query) return true;
    return `${model.display_name} ${model.model_name}`.toLowerCase().includes(query);
  });

  const handleSelect = (modelId: string) => {
    onSelect(modelId);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        选择模型
      </Button>
      <DialogContent className="max-w-4xl p-0">
        <DialogHeader className="border-b px-6 py-4">
          <DialogTitle>选择{scene.name}</DialogTitle>
          <DialogDescription>
            先选择模型厂商，再从该厂商已启用的{modelTypeLabel[scene.type] ?? scene.type}
            模型中选择。
          </DialogDescription>
        </DialogHeader>

        <div className="grid min-h-[520px] grid-cols-[280px_minmax(0,1fr)] overflow-hidden">
          <aside className="flex min-h-0 flex-col border-r bg-muted/20">
            <div className="border-b px-4 py-3 text-sm font-medium">模型厂商</div>
            <ScrollArea className="min-h-0 flex-1">
              <nav className="flex flex-col gap-1 p-3">
                {platformsWithModels.map(({ platform, models: platformModels }) => {
                  const selected = selectedPlatformId === platform.id;
                  const logo = getProviderLogo(platform);
                  return (
                    <button
                      key={platform.id}
                      type="button"
                      className={cn(
                        "flex items-center gap-3 rounded-md px-3 py-2.5 text-left outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring/50",
                        selected
                          ? "bg-background shadow-xs ring-1 ring-border"
                          : "hover:bg-background/70",
                      )}
                      onClick={() => setSelectedPlatformId(platform.id)}
                    >
                      <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-background ring-1 ring-border">
                        {logo ? (
                          <img src={logo} alt="" className="size-6 rounded-sm object-contain" />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {platform.name.slice(0, 2).toUpperCase()}
                          </span>
                        )}
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium">{platform.name}</span>
                        <span className="block text-xs text-muted-foreground">
                          {platformModels.length} 个可选模型
                        </span>
                      </span>
                      <IconChevronRight className="text-muted-foreground" />
                    </button>
                  );
                })}
              </nav>
            </ScrollArea>
          </aside>

          <section className="flex min-h-0 flex-col">
            <div className="border-b p-4">
              <div className="relative">
                <IconSearch className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  placeholder="搜索模型"
                  className="pl-9"
                />
              </div>
            </div>
            <ScrollArea className="min-h-0 flex-1">
              <div className="flex flex-col gap-2 p-4">
                {visibleModels.map((model) => {
                  const selected = selectedModelId === model.id;
                  const modelPlatform = platforms.find(
                    (platform) => platform.id === model.platform_id,
                  );
                  const logo = modelPlatform ? getProviderLogo(modelPlatform) : "";
                  return (
                    <button
                      key={model.id}
                      type="button"
                      className={cn(
                        "flex min-h-16 items-center gap-3 rounded-md border px-4 py-3 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
                        selected ? "border-primary bg-primary/5" : "hover:bg-muted/50",
                      )}
                      onClick={() => handleSelect(model.id)}
                    >
                      <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-muted/40 ring-1 ring-border">
                        {logo ? (
                          <img src={logo} alt="" className="size-6 rounded-sm object-contain" />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {(modelPlatform?.name ?? model.model_name).slice(0, 2).toUpperCase()}
                          </span>
                        )}
                      </span>
                      <span className="flex min-w-0 flex-1 flex-col">
                        <span className="truncate text-sm font-medium">
                          {model.display_name || model.model_name}
                        </span>
                        <span className="truncate font-mono text-xs text-muted-foreground">
                          {model.model_name}
                        </span>
                      </span>
                      <Badge variant="outline">
                        {modelTypeLabel[model.model_type] ?? model.model_type}
                      </Badge>
                      {selected ? <IconCheck className="text-primary" /> : null}
                    </button>
                  );
                })}

                {visibleModels.length === 0 ? (
                  <div className="flex h-40 items-center justify-center rounded-md border border-dashed text-sm text-muted-foreground">
                    没有匹配的模型
                  </div>
                ) : null}
              </div>
            </ScrollArea>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}
