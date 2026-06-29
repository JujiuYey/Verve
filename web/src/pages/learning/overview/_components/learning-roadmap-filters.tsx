import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import type { LearningRoadmap } from "@/pages/learning/mock-roadmaps";

export type LearningRoadmapFilter = "all" | LearningRoadmap["category"];

type Props = {
  value: LearningRoadmapFilter;
  onValueChange: (value: LearningRoadmapFilter) => void;
};

export function LearningRoadmapFilters({ value, onValueChange }: Props) {
  return (
    <ToggleGroup
      type="single"
      value={value}
      onValueChange={(next) => {
        if (next) onValueChange(next as LearningRoadmapFilter);
      }}
      variant="outline"
    >
      <ToggleGroupItem value="all">全部</ToggleGroupItem>
      <ToggleGroupItem value="frontend">前端实战</ToggleGroupItem>
      <ToggleGroupItem value="engineering">工程能力</ToggleGroupItem>
      <ToggleGroupItem value="ai">AI 产品</ToggleGroupItem>
    </ToggleGroup>
  );
}
