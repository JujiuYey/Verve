import { useNavigate } from "@tanstack/react-router";
import { useMemo, useState } from "react";

import { ScrollArea } from "@/components/ui/scroll-area";
import { getRoadmapList } from "@/pages/learning/mock-roadmaps";

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
  const roadmaps = getRoadmapList();

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
            先做前端页面体验，每张卡片点进去都是一条可视化学习路线。
          </p>
        </div>
        <LearningRoadmapFilters value={filter} onValueChange={setFilter} />
      </div>

      <ScrollArea className="min-h-0 flex-1 px-3">
        {visibleRoadmaps.length === 0 ? (
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
