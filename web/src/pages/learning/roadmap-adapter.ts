import type { Edge, Node } from "@xyflow/react";

import type {
  GoalDetail,
  LearningGoal,
  LearningObjective,
} from "@/api/learning/goal";

export type RoadmapStageStatus = "planned" | "active" | "completed";

export type RoadmapLesson = {
  id: string;
  title: string;
  summary: string;
  outcomes: string[];
  duration: string;
  resources: string[];
  tasks: string[];
};

export type RoadmapStage = {
  id: string;
  title: string;
  summary: string;
  description: string;
  duration: string;
  difficulty: "入门" | "进阶" | "强化";
  status: RoadmapStageStatus;
  order: number;
  x: number;
  y: number;
  outcomes: string[];
  lessons: RoadmapLesson[];
};

export type LearningRoadmap = {
  id: string;
  title: string;
  description: string;
  tagline: string;
  level: string;
  duration: string;
  progress: number;
  learners: string;
  category: "frontend" | "engineering" | "ai";
  tags: string[];
  heroPoints: string[];
  stages: RoadmapStage[];
};

type RoadmapNodeData = {
  label: string;
  duration: string;
  difficulty: string;
  status: RoadmapStageStatus;
};

export function goalToRoadmap(goal: LearningGoal): LearningRoadmap {
  return {
    id: goal.id,
    title: goal.title,
    description: displayDescription(goal),
    tagline: goal.status === "active" ? "正在学习的路线" : statusLabel(goal.status),
    level: goal.source === "documents" ? "文档生成" : "自主目标",
    duration: "持续推进",
    progress: goal.status === "completed" ? 100 : 0,
    learners: "个人学习",
    category: goal.source === "documents" ? "engineering" : "ai",
    tags: [sourceLabel(goal.source), statusLabel(goal.status)],
    heroPoints: [],
    stages: [],
  };
}

export function goalDetailToRoadmap(detail: GoalDetail): LearningRoadmap {
  const objectives = detail.objectives || [];
  const stages = objectivesToStages(objectives, detail.current_objective_id);
  const progress = detail.progress
    ? Math.round((detail.progress.completed / Math.max(detail.progress.total, 1)) * 100)
    : 0;

  return {
    ...goalToRoadmap(detail.goal),
    description: displayDescription(detail.goal),
    tagline: detail.path?.status === "active" ? "按阶段推进，每次聚焦一个小目标" : statusLabel(detail.goal.status),
    duration: `${stages.length} 个阶段`,
    progress,
    tags: [sourceLabel(detail.goal.source), `${objectives.length} 个小目标`],
    stages,
  };
}

function displayDescription(goal: LearningGoal) {
  if (goal.source === "documents") return sourceDescription(goal.source);
  return goal.description || sourceDescription(goal.source);
}

export function getRoadmapFlow(roadmap: LearningRoadmap): {
  nodes: Node<RoadmapNodeData>[];
  edges: Edge[];
} {
  const nodes = roadmap.stages.map((stage) => ({
    id: stage.id,
    type: "roadmapNode",
    position: { x: stage.x, y: stage.y },
    data: {
      label: stage.title,
      duration: stage.duration,
      difficulty: stage.difficulty,
      status: stage.status,
    },
  }));

  const edges = roadmap.stages.slice(1).map((stage, index) => ({
    id: `${roadmap.stages[index]?.id}-${stage.id}`,
    source: roadmap.stages[index]?.id ?? stage.id,
    target: stage.id,
    animated: stage.status === "active",
    style: {
      stroke: stage.status === "completed" ? "var(--primary)" : "var(--border)",
      strokeWidth: 2,
    },
  }));

  return { nodes, edges };
}

function objectivesToStages(
  objectives: LearningObjective[],
  currentObjectiveId?: string | null,
): RoadmapStage[] {
  const grouped = new Map<string, LearningObjective[]>();
  const orderedObjectives = [...objectives].sort((a, b) => a.order_index - b.order_index);

  for (const objective of orderedObjectives) {
    const stageTitle = objective.stage_title || "学习阶段";
    const stageObjectives = grouped.get(stageTitle) || [];
    stageObjectives.push(objective);
    grouped.set(stageTitle, stageObjectives);
  }

  return Array.from(grouped.entries()).map(([title, stageObjectives], index) => {
    const completedCount = stageObjectives.filter((item) => item.status === "completed").length;
    const hasActive =
      stageObjectives.some((item) => item.status === "active" || item.id === currentObjectiveId) ||
      (index === 0 && !currentObjectiveId && completedCount < stageObjectives.length);
    const status: RoadmapStageStatus =
      completedCount === stageObjectives.length ? "completed" : hasActive ? "active" : "planned";

    return {
      id: stageObjectives[0]?.id || `${index}`,
      title,
      summary: `${stageObjectives.length} 个小目标`,
      description: stageObjectives.map((item) => item.detail || item.title).join("\n"),
      duration: `阶段 ${index + 1}`,
      difficulty: difficultyForIndex(index),
      status,
      order: index + 1,
      x: 70 + index * 280,
      y: index % 2 === 0 ? 90 : 30,
      outcomes: stageObjectives.map((item) => item.title),
      lessons: stageObjectives.map((item, lessonIndex) => ({
        id: item.id,
        title: item.title,
        summary: item.detail || "围绕这个小目标完成理解、解释和练习。",
        outcomes: [masteryLabel(item.mastery_level), statusLabel(item.status)],
        duration: `小目标 ${lessonIndex + 1}`,
        resources: [item.stage_title || title],
        tasks: ["用自己的话解释这个知识点", "完成一次练习并记录学习日志"],
      })),
    };
  });
}

function difficultyForIndex(index: number): RoadmapStage["difficulty"] {
  if (index === 0) return "入门";
  if (index <= 2) return "进阶";
  return "强化";
}

function sourceDescription(source: string) {
  if (source === "documents") return "基于文档管理中的资料生成的学习路线。";
  return "基于一句话学习目标生成的学习路线。";
}

function sourceLabel(source: string) {
  if (source === "documents") return "文档生成";
  return "目标生成";
}

function statusLabel(status: string) {
  const labels: Record<string, string> = {
    active: "进行中",
    archived: "已归档",
    completed: "已完成",
    pending: "待开始",
    review: "复习中",
  };
  return labels[status] || status;
}

function masteryLabel(level: string) {
  const labels: Record<string, string> = {
    none: "尚未验证",
    seen: "已看过",
    heard: "已听过",
    explained: "能解释",
    written: "能写出",
    verified: "已验证",
  };
  return labels[level] || level;
}
