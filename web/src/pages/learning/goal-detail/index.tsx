import { useNavigate, useParams } from "@tanstack/react-router";
import { ReactFlowProvider } from "@xyflow/react";
import { useEffect, useMemo, useState } from "react";

import { useDeleteGoal, useGoalDetail } from "@/api/learning/goal";
import { getRoadmapFlow, goalDetailToRoadmap } from "@/pages/learning/roadmap-adapter";

import { LearningRoadmapDetailHeader } from "./_components/learning-roadmap-detail-header";
import { LearningRoadmapDetailPanels } from "./_components/learning-roadmap-detail-panels";
import { LearningRoadmapFlow } from "./_components/learning-roadmap-flow";

export function GoalDetailPage() {
  const navigate = useNavigate();
  const { goalId } = useParams({ from: "/_layout/learn/goal/$goalId" });
  const { data, isLoading } = useGoalDetail(goalId);
  const deleteGoal = useDeleteGoal();
  const roadmap = useMemo(() => (data ? goalDetailToRoadmap(data) : null), [data]);

  const [activeStageId, setActiveStageId] = useState(roadmap?.stages[0]?.id ?? "");
  const [collapsedFolderPaths, setCollapsedFolderPaths] = useState<Set<string>>(new Set());

  useEffect(() => {
    setActiveStageId(roadmap?.stages[0]?.id ?? "");
    setCollapsedFolderPaths(new Set());
  }, [roadmap?.id, roadmap?.stages]);

  const flow = useMemo(
    () => (roadmap ? getRoadmapFlow(roadmap, collapsedFolderPaths) : { nodes: [], edges: [] }),
    [collapsedFolderPaths, roadmap],
  );

  const activeStage =
    roadmap?.stages.find((stage) => stage.id === activeStageId) ?? roadmap?.stages[0] ?? null;

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">正在加载学习项目...</div>;
  }

  if (!roadmap) {
    return <div className="p-6 text-sm text-muted-foreground">这个学习项目不存在</div>;
  }

  return (
    <ReactFlowProvider>
      <div className="flex h-full flex-col gap-6 overflow-hidden p-6">
        <LearningRoadmapDetailHeader
          roadmap={roadmap}
          isDeleting={deleteGoal.isPending}
          onBack={() => navigate({ to: "/" })}
          onDelete={() => {
            deleteGoal.mutate(goalId, {
              onSuccess: () => navigate({ to: "/" }),
            });
          }}
        />

        <div className="grid min-h-0 flex-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_380px]">
          <LearningRoadmapFlow
            nodes={flow.nodes}
            edges={flow.edges}
            onStageClick={(stageId) => setActiveStageId(stageId)}
            onFolderToggle={(folderPath) => {
              setCollapsedFolderPaths((prev) => {
                const next = new Set(prev);
                if (next.has(folderPath)) {
                  next.delete(folderPath);
                } else {
                  next.add(folderPath);
                }
                return next;
              });
            }}
            onObjectiveClick={(objectiveId) =>
              navigate({
                to: "/learn/feynman-practice/$goalId/$objectiveId",
                params: { goalId, objectiveId },
              })
            }
          />
          <LearningRoadmapDetailPanels goalId={goalId} activeStage={activeStage} />
        </div>
      </div>
    </ReactFlowProvider>
  );
}
