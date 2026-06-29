import {
  BookOpenIcon,
  BrainCircuitIcon,
  Clock3Icon,
  LayersIcon,
  TrendingUpIcon,
  UsersIcon,
  WorkflowIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import { type LearningRoadmap } from "@/pages/learning/mock-roadmaps";

const categoryMeta: Record<
  LearningRoadmap["category"],
  { label: string; icon: typeof WorkflowIcon; className: string }
> = {
  frontend: {
    label: "前端实战",
    icon: BookOpenIcon,
    className: "bg-sky-500/10 text-sky-700 dark:text-sky-300",
  },
  engineering: {
    label: "工程能力",
    icon: WorkflowIcon,
    className: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  },
  ai: {
    label: "AI 产品",
    icon: BrainCircuitIcon,
    className: "bg-amber-500/10 text-amber-700 dark:text-amber-300",
  },
};

type Props = {
  roadmaps: LearningRoadmap[];
  onOpenRoadmap: (roadmap: LearningRoadmap) => void;
};

export function LearningRoadmapGrid({ roadmaps, onOpenRoadmap }: Props) {
  return (
    <div className="grid grid-cols-1 gap-3 xl:grid-cols-2 2xl:grid-cols-3">
      {roadmaps.map((roadmap) => {
        const meta = categoryMeta[roadmap.category];
        const Icon = meta.icon;

        return (
          <Card
            key={roadmap.id}
            className="rounded-2xl py-0 transition-colors hover:border-primary/50"
          >
            <CardHeader className="gap-3 border-b p-4!">
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-2.5">
                  <Badge variant="secondary" className={cn("w-fit", meta.className)}>
                    <Icon />
                    {meta.label}
                  </Badge>
                  <div className="space-y-1.5">
                    <CardTitle className="text-xl leading-7">{roadmap.title}</CardTitle>
                    <CardDescription className="leading-6">{roadmap.description}</CardDescription>
                  </div>
                </div>
                <Badge variant="outline">{roadmap.level}</Badge>
              </div>
            </CardHeader>

            <CardContent className="space-y-4">
              <p className="text-sm leading-6 text-muted-foreground">{roadmap.tagline}</p>

              <div className="grid grid-cols-2 gap-2.5">
                <InfoBlock icon={Clock3Icon} label="周期" value={roadmap.duration} />
                <InfoBlock icon={TrendingUpIcon} label="进度" value={`${roadmap.progress}%`} />
                <InfoBlock icon={LayersIcon} label="阶段数" value={`${roadmap.stages.length} 个`} />
                <InfoBlock icon={UsersIcon} label="学习人数" value={roadmap.learners} />
              </div>

              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">整体进展</span>
                  <span className="font-medium">{roadmap.progress}%</span>
                </div>
                <Progress value={roadmap.progress} />
              </div>

              <div className="flex flex-wrap gap-2">
                {roadmap.tags.map((tag) => (
                  <Badge key={tag} variant="outline">
                    {tag}
                  </Badge>
                ))}
              </div>
            </CardContent>

            <CardFooter className="justify-between border-t py-3">
              <div className="text-sm text-muted-foreground">点击后进入路线图详情和节点说明</div>
              <Button onClick={() => onOpenRoadmap(roadmap)}>打开学习地图</Button>
            </CardFooter>
          </Card>
        );
      })}
    </div>
  );
}

function InfoBlock({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Clock3Icon;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-xl border bg-muted/30 px-3 py-2.5">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Icon className="size-3.5" />
        {label}
      </div>
      <div className="mt-0.5 text-sm font-medium">{value}</div>
    </div>
  );
}
