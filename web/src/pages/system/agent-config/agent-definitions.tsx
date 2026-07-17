import {
  IconChecks,
  IconChalkboardTeacher,
  IconEdit,
  IconMessageQuestion,
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
    name: "知识问答",
    shortName: "问答",
    description: "检索全部 Wiki，并结合相关学习记录生成结构化回答。",
    runtime: "KnowledgeQAService",
    icon: IconMessageQuestion,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        required: true,
        description: "用于生成知识回答和学习建议。",
      },
    ],
  },
  {
    key: "learning_teacher",
    name: "Feynman 教师",
    shortName: "教师",
    description: "基于 Wiki 证据以费曼方式解释知识点的对话 Agent。",
    runtime: "LearningTeacher",
    icon: IconChalkboardTeacher,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        required: true,
        description: "用于生成费曼式讲解与教学追问。",
      },
    ],
  },
  {
    key: "feynman_reviewer",
    name: "Feynman 复盘",
    shortName: "复盘",
    description: "倾听学习者解释并返回结构化复盘意见。",
    runtime: "FeynmanReviewer",
    icon: IconChecks,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        required: true,
        description: "用于结构化分析学习者解释并给出复盘意见。",
      },
    ],
  },
  {
    key: "wiki_curator",
    name: "Wiki 编辑",
    shortName: "编辑",
    description: "根据用户指令提出完整的 Wiki Markdown 修改建议。",
    runtime: "WikiCurator",
    icon: IconEdit,
    scenes: [
      {
        key: "default",
        name: "默认模型",
        required: true,
        description: "用于生成 Wiki Markdown 修改提案。",
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
