import { useNavigate } from "@tanstack/react-router";
import { AlertCircleIcon } from "lucide-react";
import { useMemo } from "react";

import { useGoalList } from "@/api/learning/goal";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { goalToRoadmap } from "@/pages/learning/roadmap-adapter";

import { LearningRoadmapEmptyState } from "./_components/learning-roadmap-empty-state";
import { LearningRoadmapGrid } from "./_components/learning-roadmap-grid";

export function LearningOverviewPage() {
  const navigate = useNavigate();
  const { data, error, isError, isLoading, refetch } = useGoalList(1, 50);
  const roadmaps = useMemo(() => (data?.data || []).map(goalToRoadmap), [data?.data]);

  return (
    <div className="flex h-full flex-col gap-6 p-6">
      <div className="flex flex-col gap-1 px-1">
        <h2 className="text-xl font-semibold">挑一个方向继续</h2>
        <p className="text-sm text-muted-foreground">
          每个方向都基于你的学习目标，按阶段推进。点开卡片看完整路线和当前该做的小目标。
        </p>
      </div>

      <ScrollArea className="min-h-0 flex-1 px-3">
        {isLoading ? (
          <div className="text-sm text-muted-foreground">正在加载你的学习方向…</div>
        ) : isError ? (
          <Alert variant="destructive" className="rounded-2xl">
            <AlertCircleIcon />
            <AlertTitle>暂时拿不到你的学习方向</AlertTitle>
            <AlertDescription className="flex flex-col gap-3">
              <span>
                {error instanceof Error
                  ? error.message
                  : "可能是网络不太稳，稍后再试一次，或者刷新页面看看。"}
              </span>
              <Button variant="outline" className="w-fit" onClick={() => refetch()}>
                再试一次
              </Button>
            </AlertDescription>
          </Alert>
        ) : roadmaps.length === 0 ? (
          <LearningRoadmapEmptyState />
        ) : (
          <LearningRoadmapGrid
            roadmaps={roadmaps}
            onOpenRoadmap={(roadmap) => {
              navigate({ to: "/learn/goal/$goalId", params: { goalId: roadmap.id } });
            }}
          />
        )}
      </ScrollArea>
    </div>
  );
}
