import { Background, Controls, type Edge, type Node, type NodeProps } from "@xyflow/react";

import { Canvas } from "@/components/ai-elements/canvas";

import { LearningRoadmapNode } from "./learning-roadmap-node";

type RoadmapNodeData = {
  label: string;
  duration: string;
  difficulty: string;
  status: "planned" | "active" | "completed";
};

const nodeTypes = {
  roadmapNode: LearningRoadmapNode as React.ComponentType<NodeProps>,
};

type Props = {
  nodes: Node<RoadmapNodeData>[];
  edges: Edge[];
  onNodeClick: (stageId: string) => void;
};

export function LearningRoadmapFlow({ nodes, edges, onNodeClick }: Props) {
  return (
    <div className="h-full overflow-hidden rounded-2xl border bg-background py-0">
      <Canvas
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodeClick={(_, node) => onNodeClick(node.id)}
        nodesDraggable={false}
        elementsSelectable
        fitViewOptions={{ padding: 0.2 }}
      >
        <Controls showInteractive={false} />
        <Background gap={20} color="var(--border)" />
      </Canvas>
    </div>
  );
}
