import {
  IconBook2,
  IconChecklist,
  IconRoute,
  IconStack2,
  IconTargetArrow,
  IconUserQuestion,
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
        required: true,
        description: "解析知识库时生成文档块向量；未配置时不会创建解析任务。",
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
