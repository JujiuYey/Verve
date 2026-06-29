import { ArrowLeftIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import type { LearningRoadmap } from "@/pages/learning/mock-roadmaps";

type Props = {
  roadmap: LearningRoadmap;
  onBack: () => void;
};

export function LearningRoadmapDetailHeader({ roadmap, onBack }: Props) {
  return (
    <section className="w-full flex flex-col gap-4 rounded-2xl border bg-background p-4">
      <div className="flex justify-between gap-2">
        <Button variant="secondary" className="w-fit px-0" onClick={onBack}>
          <ArrowLeftIcon data-icon="inline-start" />
          返回学习项目
        </Button>
        <div className="flex flex-wrap items-center justify-end gap-2">
          <Badge variant="outline">{roadmap.level}</Badge>
          <Badge variant="secondary">{roadmap.duration}</Badge>
          <Badge variant="outline">{roadmap.learners}</Badge>
        </div>
      </div>

      <div>
        <h1 className="text-3xl font-semibold tracking-tight">{roadmap.title}</h1>
        <p className="max-w-3xl text-sm leading-6 text-muted-foreground">{roadmap.description}</p>
      </div>
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">整体学习进展</span>
          <span className="font-medium">{roadmap.progress}%</span>
        </div>
        <Progress value={roadmap.progress} />
      </div>
    </section>
  );
}
