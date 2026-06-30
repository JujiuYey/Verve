import { useNavigate } from "@tanstack/react-router";
import { AlertCircleIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { useGoalList } from "@/api/learning/goal";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { goalToRoadmap } from "@/pages/learning/roadmap-adapter";

import { LearningOverviewHeader } from "./_components/learning-overview-header";
import { LearningRoadmapEmptyState } from "./_components/learning-roadmap-empty-state";
import {
  LearningRoadmapFilters,
  type LearningRoadmapFilter,
} from "./_components/learning-roadmap-filters";
import { LearningRoadmapGrid } from "./_components/learning-roadmap-grid";

type FilterValue = LearningRoadmapFilter;

export function LearningOverviewPage() {
  const navigate = useNavigate();
  const [filter, setFilter] = useState<FilterValue>("all");
  const { data, error, isError, isLoading, refetch } = useGoalList(1, 50);
  const roadmaps = useMemo(() => (data?.data || []).map(goalToRoadmap), [data?.data]);

  const visibleRoadmaps = useMemo(() => {
    if (filter === "all") return roadmaps;
    return roadmaps.filter((roadmap) => roadmap.category === filter);
  }, [filter, roadmaps]);

  const totalStages = roadmaps.reduce((sum, item) => sum + item.stages.length, 0);

  return (
    <div className="flex h-full flex-col gap-6 p-6">
      <LearningOverviewHeader roadmapCount={roadmaps.length} stageCount={totalStages} />

      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 className="text-xl font-semibold">学习项目列表</h2>
          <p className="text-sm text-muted-foreground">
            从真实学习目标进入路线图详情，继续推进当前阶段的小目标。
          </p>
        </div>
        <LearningRoadmapFilters value={filter} onValueChange={setFilter} />
      </div>

      <ScrollArea className="min-h-0 flex-1 px-3">
        {isLoading ? (
          <div className="text-sm text-muted-foreground">正在加载学习项目...</div>
        ) : isError ? (
          <Alert variant="destructive" className="rounded-2xl">
            <AlertCircleIcon />
            <AlertTitle>学习目标接口读取失败</AlertTitle>
            <AlertDescription className="flex flex-col gap-3">
              <span>
                {error instanceof Error
                  ? error.message
                  : "请查看后端日志里的学习目标分页查询信息。"}
              </span>
              <Button variant="outline" className="w-fit" onClick={() => refetch()}>
                重新加载
              </Button>
            </AlertDescription>
          </Alert>
        ) : visibleRoadmaps.length === 0 ? (
          <LearningRoadmapEmptyState onReset={() => setFilter("all")} />
        ) : (
          <LearningRoadmapGrid
            roadmaps={visibleRoadmaps}
            onOpenRoadmap={(roadmap) => {
              navigate({ to: "/learn/goal/$goalId", params: { goalId: roadmap.id } });
            }}
          />
        )}
      </ScrollArea>
    </div>
  );
}
