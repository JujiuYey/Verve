import { SparklesIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

type Props = {
  onReset: () => void;
};

export function LearningRoadmapEmptyState({ onReset }: Props) {
  return (
    <Empty className="min-h-64 rounded-2xl border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <SparklesIcon />
        </EmptyMedia>
        <EmptyTitle>这个分类还没有项目</EmptyTitle>
        <EmptyDescription>先切回全部，或者下一步我帮你补一个新的学习方向。</EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="outline" onClick={onReset}>
          查看全部项目
        </Button>
      </EmptyContent>
    </Empty>
  );
}
