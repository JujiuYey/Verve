import type { Edge, Node } from "@xyflow/react";

import type { GoalDetail, LearningGoal, LearningObjective } from "@/api/learning/goal";

export type RoadmapStageStatus = "planned" | "active" | "completed";

export type RoadmapLesson = {
  id: string;
  title: string;
  summary: string;
  sourceFolderPath?: string;
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
  kind: "folder" | "objective";
  stageId: string;
  folderPath?: string;
  isCollapsed?: boolean;
  childCount?: number;
  objectiveId?: string;
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
    tagline:
      detail.path?.status === "active"
        ? "按阶段推进，每次聚焦一个小目标"
        : statusLabel(detail.goal.status),
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

export function getRoadmapFlow(
  roadmap: LearningRoadmap,
  collapsedFolderPaths: Set<string> = new Set(),
): {
  nodes: Node<RoadmapNodeData>[];
  edges: Edge[];
} {
  const layoutItems = roadmap.stages.flatMap((stage) =>
    stage.lessons.map((lesson) => ({
      stage,
      lesson,
      pathParts: normalizeFolderPath(lesson.sourceFolderPath || stage.title),
    })),
  );
  const branchColumns = assignBranchColumns(layoutItems.map((item) => item.pathParts));
  const directorySlots = assignDirectorySlots(layoutItems.map((item) => item.pathParts));
  const nodes: Node<RoadmapNodeData>[] = [];
  const edges: Edge[] = [];
  const folderNodeIDs = new Set<string>();
  const lessonIndexesByDirectory = new Map<string, number>();

  for (const item of layoutItems) {
    const { stage, lesson, pathParts } = item;
    if (hasCollapsedAncestor(pathParts, collapsedFolderPaths)) continue;

    const column = branchColumns.get(branchKeyForPath(pathParts)) ?? 0;
    let parentNodeID = "";

    pathParts.forEach((part, depth) => {
      const folderPath = pathParts.slice(0, depth + 1).join(" / ");
      const folderNodeID = `folder:${folderPath}`;
      if (!folderNodeIDs.has(folderNodeID)) {
        folderNodeIDs.add(folderNodeID);
        const folderColumn = depth === 0 ? 0 : column;
        const slot = directorySlots.get(folderPath) ?? 0;
        nodes.push({
          id: folderNodeID,
          type: "roadmapNode",
          position: { x: 70 + folderColumn * 320, y: 30 + depth * 150 + slot * 110 },
          data: {
            label: part,
            duration: depth === 0 ? "根目录" : `第 ${depth + 1} 层目录`,
            difficulty: "文件夹",
            status: stage.status,
            kind: "folder",
            stageId: stage.id,
            folderPath,
            isCollapsed: collapsedFolderPaths.has(folderPath),
            childCount: childCountForFolder(
              folderPath,
              layoutItems.map((entry) => entry.pathParts),
            ),
          },
        });
      }

      if (parentNodeID) {
        edges.push(buildRoadmapEdge(parentNodeID, folderNodeID, stage.status));
      }
      parentNodeID = folderNodeID;
    });

    if (collapsedFolderPaths.has(pathParts.join(" / "))) continue;

    const directoryPath = pathParts.join(" / ");
    const lessonIndex = lessonIndexesByDirectory.get(directoryPath) ?? 0;
    lessonIndexesByDirectory.set(directoryPath, lessonIndex + 1);
    const directorySlot = directorySlots.get(directoryPath) ?? 0;

    nodes.push({
      id: lesson.id,
      type: "roadmapNode",
      position: {
        x: 70 + column * 320,
        y: 30 + pathParts.length * 150 + directorySlot * 110 + lessonIndex * 132,
      },
      data: {
        label: lesson.title,
        duration: lesson.duration,
        difficulty: "小目标",
        status: stage.status,
        kind: "objective",
        stageId: stage.id,
        objectiveId: lesson.id,
      },
    });
    edges.push(buildRoadmapEdge(parentNodeID, lesson.id, stage.status));
  }

  return { nodes, edges: uniqueEdges(edges) };
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
        sourceFolderPath: item.source_folder_path,
        outcomes: [masteryLabel(item.mastery_level), statusLabel(item.status)],
        duration: `小目标 ${lessonIndex + 1}`,
        resources: [item.source_folder_path || item.stage_title || title],
        tasks: ["用自己的话解释这个知识点", "完成一次练习并记录学习日志"],
      })),
    };
  });
}

function normalizeFolderPath(path: string) {
  return path
    .split("/")
    .map((part) => part.trim())
    .filter(Boolean);
}

function branchKeyForPath(pathParts: string[]) {
  return pathParts.length > 1 ? pathParts.slice(0, 2).join(" / ") : pathParts[0] || "学习资料";
}

function assignBranchColumns(paths: string[][]) {
  const columns = new Map<string, number>();
  for (const pathParts of paths) {
    const branchKey = branchKeyForPath(pathParts);
    if (!columns.has(branchKey)) {
      columns.set(branchKey, columns.size);
    }
  }
  return columns;
}

function assignDirectorySlots(paths: string[][]) {
  const slots = new Map<string, number>();
  for (const pathParts of paths) {
    for (let depth = 0; depth < pathParts.length; depth++) {
      const directoryPath = pathParts.slice(0, depth + 1).join(" / ");
      if (!slots.has(directoryPath)) {
        slots.set(directoryPath, slots.size);
      }
    }
  }
  return slots;
}

function hasCollapsedAncestor(pathParts: string[], collapsedFolderPaths: Set<string>) {
  for (let depth = 0; depth < pathParts.length - 1; depth++) {
    if (collapsedFolderPaths.has(pathParts.slice(0, depth + 1).join(" / "))) {
      return true;
    }
  }
  return false;
}

function childCountForFolder(folderPath: string, paths: string[][]) {
  const directChildren = new Set<string>();
  for (const pathParts of paths) {
    const parts = folderPath.split(" / ");
    const isDescendant = parts.every((part, index) => pathParts[index] === part);
    if (!isDescendant) continue;

    if (pathParts.length === parts.length) {
      directChildren.add(`objective:${pathParts.join(" / ")}`);
    } else {
      directChildren.add(`folder:${pathParts.slice(0, parts.length + 1).join(" / ")}`);
    }
  }
  return directChildren.size;
}

function buildRoadmapEdge(source: string, target: string, status: RoadmapStageStatus): Edge {
  return {
    id: `${source}-${target}`,
    source,
    target,
    animated: status === "active",
    style: {
      stroke: status === "completed" ? "var(--primary)" : "var(--border)",
      strokeWidth: 2,
    },
  };
}

function uniqueEdges(edges: Edge[]) {
  const seen = new Set<string>();
  return edges.filter((edge) => {
    if (seen.has(edge.id)) return false;
    seen.add(edge.id);
    return true;
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
