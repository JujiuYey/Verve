import { SparklesIcon } from "lucide-react";

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export function LearningRoadmapEmptyState() {
  return (
    <Empty className="min-h-64 rounded-2xl border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <SparklesIcon />
        </EmptyMedia>
        <EmptyTitle>这里还没有学习方向</EmptyTitle>
        <EmptyDescription>
          写下第一个想学的东西，Verve 会帮你拆成阶段和小目标，每一步都知道下一步要做什么。
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}
