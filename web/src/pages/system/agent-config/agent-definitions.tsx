import {
  IconRoute,
  IconStack2,
} from "@tabler/icons-react";
import type { ComponentType, SVGProps } from "react";

export type SceneDefinition = {
  key: string;
  name: string;
  required?: boolean;
  description: string;
};

export type AgentDefinition = {
  key: string;
  name: string;
  shortName: string;
  description: string;
  runtime: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  scenes: SceneDefinition[];
};

export type AgentStatus = "ready" | "partial" | "missing";

export type SceneConfigState = {
  model_id: string;
  enabled: boolean;
};

export const AGENTS: AgentDefinition[] = [
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
        required: true,
        description: "用于对话、工具调用决策和下一步动作规划。",
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
        required: true,
        description: "上传或保存文档时生成文档块向量；未配置时不会创建解析任务。",
      },
    ],
  },
];

export function getSceneKey(agentKey: string, sceneKey: string) {
  return `${agentKey}.${sceneKey}`;
}

export function getAgentStatus(
  agent: AgentDefinition,
  configsByScene: Map<string, SceneConfigState>,
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
