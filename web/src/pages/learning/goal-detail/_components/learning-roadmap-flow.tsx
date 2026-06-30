import { Background, Controls, type Edge, type Node, type NodeProps } from "@xyflow/react";

import { Canvas } from "@/components/ai-elements/canvas";

import { LearningRoadmapNode } from "./learning-roadmap-node";

type RoadmapNodeData = {
  label: string;
  duration: string;
  difficulty: string;
  status: "planned" | "active" | "completed";
  kind: "folder" | "objective";
  stageId: string;
  folderPath?: string;
  isCollapsed?: boolean;
  childCount?: number;
  objectiveId?: string;
};

const nodeTypes = {
  roadmapNode: LearningRoadmapNode as React.ComponentType<NodeProps>,
};

type Props = {
  nodes: Node<RoadmapNodeData>[];
  edges: Edge[];
  onStageClick: (stageId: string) => void;
  onFolderToggle: (folderPath: string) => void;
  onObjectiveClick: (objectiveId: string) => void;
};

export function LearningRoadmapFlow({
  nodes,
  edges,
  onStageClick,
  onFolderToggle,
  onObjectiveClick,
}: Props) {
  return (
    <div className="h-full overflow-hidden rounded-2xl border bg-background py-0">
      <Canvas
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodeClick={(_, node) => {
          const roadmapNode = node as Node<RoadmapNodeData>;
          if (roadmapNode.data.kind === "objective" && roadmapNode.data.objectiveId) {
            onObjectiveClick(roadmapNode.data.objectiveId);
            return;
          }
          if (roadmapNode.data.kind === "folder" && roadmapNode.data.folderPath) {
            onFolderToggle(roadmapNode.data.folderPath);
          }
          onStageClick(roadmapNode.data.stageId);
        }}
        nodesDraggable={false}
        panOnDrag
        selectionOnDrag={false}
        elementsSelectable
        minZoom={0.15}
        fitViewOptions={{ padding: 0.08 }}
      >
        <Controls showInteractive={false} />
        <Background gap={20} color="var(--border)" />
      </Canvas>
    </div>
  );
}
