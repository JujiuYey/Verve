import { useNavigate, useParams } from "@tanstack/react-router";
import { ReactFlowProvider } from "@xyflow/react";
import { useMemo, useState } from "react";

import { getRoadmapById, getRoadmapFlow } from "@/pages/learning/mock-roadmaps";

import { LearningRoadmapDetailHeader } from "./_components/learning-roadmap-detail-header";
import { LearningRoadmapDetailPanels } from "./_components/learning-roadmap-detail-panels";
import { LearningRoadmapFlow } from "./_components/learning-roadmap-flow";

export function GoalDetailPage() {
  const navigate = useNavigate();
  const { goalId } = useParams({ from: "/_layout/learn/goal/$goalId" });
  const roadmap = getRoadmapById(goalId);

  const [activeStageId, setActiveStageId] = useState(roadmap?.stages[0]?.id ?? "");

  const flow = useMemo(
    () => (roadmap ? getRoadmapFlow(roadmap) : { nodes: [], edges: [] }),
    [roadmap],
  );

  const activeStage =
    roadmap?.stages.find((stage) => stage.id === activeStageId) ?? roadmap?.stages[0] ?? null;

  if (!roadmap) {
    return <div className="p-6 text-sm text-muted-foreground">这条学习路线不存在</div>;
  }

  return (
    <ReactFlowProvider>
      <div className="flex h-full flex-col gap-6 overflow-hidden p-6">
        <LearningRoadmapDetailHeader
          roadmap={roadmap}
          onBack={() => navigate({ to: "/" })}
        />

        <div className="grid min-h-0 flex-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_380px]">
          <LearningRoadmapFlow
            nodes={flow.nodes}
            edges={flow.edges}
            onNodeClick={(stageId) => setActiveStageId(stageId)}
          />
          <LearningRoadmapDetailPanels activeStage={activeStage} />
        </div>
      </div>
    </ReactFlowProvider>
  );
}
