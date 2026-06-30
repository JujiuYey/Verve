import { useNavigate, useParams } from "@tanstack/react-router";
import { ReactFlowProvider } from "@xyflow/react";
import { Loader2Icon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { useGoalDetail } from "@/api/learning/goal";
import { Card, CardContent } from "@/components/ui/card";
import { getRoadmapFlow, goalDetailToRoadmap } from "@/pages/learning/roadmap-adapter";

import { LearningRoadmapDetailHeader } from "./_components/learning-roadmap-detail-header";
import { LearningRoadmapDetailPanels } from "./_components/learning-roadmap-detail-panels";
import { LearningRoadmapFlow } from "./_components/learning-roadmap-flow";

export function GoalDetailPage() {
  const navigate = useNavigate();
  const { goalId } = useParams({ from: "/_layout/learn/goal/$goalId" });
  const { data, isLoading } = useGoalDetail(goalId);
  const roadmap = useMemo(() => (data ? goalDetailToRoadmap(data) : null), [data]);

  const [activeStageId, setActiveStageId] = useState(roadmap?.stages[0]?.id ?? "");

  useEffect(() => {
    setActiveStageId(roadmap?.stages[0]?.id ?? "");
  }, [roadmap?.id, roadmap?.stages]);

  const flow = useMemo(
    () => (roadmap ? getRoadmapFlow(roadmap) : { nodes: [], edges: [] }),
    [roadmap],
  );

  const activeStage =
    roadmap?.stages.find((stage) => stage.id === activeStageId) ?? roadmap?.stages[0] ?? null;
  const isGenerating = !!data?.goal && !data.path;

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">正在加载学习路线...</div>;
  }

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

        {isGenerating ? (
          <Card className="min-h-80 rounded-2xl">
            <CardContent className="flex h-full min-h-80 flex-col items-center justify-center gap-3 text-center">
              <Loader2Icon className="size-8 animate-spin text-primary" />
              <div className="text-base font-medium">学习路线正在生成中</div>
              <div className="max-w-md text-sm leading-6 text-muted-foreground">
                学习目标已经保存，后台正在根据资料拆解阶段和小目标。页面会自动刷新生成结果。
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="grid min-h-0 flex-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_380px]">
            <LearningRoadmapFlow
              nodes={flow.nodes}
              edges={flow.edges}
              onNodeClick={(stageId) => setActiveStageId(stageId)}
            />
            <LearningRoadmapDetailPanels activeStage={activeStage} />
          </div>
        )}
      </div>
    </ReactFlowProvider>
  );
}
