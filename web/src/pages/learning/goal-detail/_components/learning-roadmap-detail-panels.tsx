import {
  BookMarkedIcon,
  CheckCircle2Icon,
  CircleDotIcon,
  Clock3Icon,
  PlayCircleIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { RoadmapStage } from "@/pages/learning/roadmap-adapter";

type Props = {
  activeStage: RoadmapStage | null;
};

export function LearningRoadmapDetailPanels({ activeStage }: Props) {
  return (
    <Card className="min-h-0 rounded-2xl py-0">
      <CardHeader className="border-b p-4!">
        <div className="flex items-start justify-between gap-3">
          <div className="space-y-2">
            <CardTitle>{activeStage?.title ?? "选择一个节点"}</CardTitle>
            <div className="text-sm text-muted-foreground">
              {activeStage?.summary ?? "点击左侧节点后，这里显示阶段说明和课程清单。"}
            </div>
          </div>
          {activeStage ? <Badge variant="secondary">{activeStage.difficulty}</Badge> : null}
        </div>
      </CardHeader>
      <CardContent className="min-h-0 p-0 pb-2">
        <ScrollArea className="h-full">
          {activeStage ? <StagePanel stage={activeStage} /> : null}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

function StagePanel({ stage }: { stage: RoadmapStage }) {
  return (
    <div className="flex flex-col gap-5 px-3">
      {/* <div className="rounded-xl bg-muted/60 p-4">
        <div className="mb-2 flex items-center gap-2 text-sm font-medium">
          <FlagIcon className="size-4" />
          这一阶段在做什么
        </div>
        <p className="text-sm leading-6 text-muted-foreground">{stage.description}</p>
      </div> */}

      <div className="grid grid-cols-2 gap-3">
        <StageMetaCard icon={Clock3Icon} label="预计周期" value={stage.duration} />
        <StageMetaCard icon={BookMarkedIcon} label="难度" value={stage.difficulty} />
      </div>

      <SectionTitle title="阶段目标" />
      <div className="flex flex-col gap-2">
        {stage.outcomes.map((outcome) => (
          <div key={outcome} className="flex items-start gap-2 text-sm text-muted-foreground">
            <CheckCircle2Icon className="mt-0.5 size-4 text-emerald-500" />
            <span>{outcome}</span>
          </div>
        ))}
      </div>

      <SectionTitle title="课程卡片" />
      <div className="flex flex-col gap-3">
        {stage.lessons.map((lesson) => (
          <div key={lesson.id} className="rounded-xl border p-4">
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1">
                <div className="font-medium">{lesson.title}</div>
                <div className="text-sm leading-6 text-muted-foreground">{lesson.summary}</div>
              </div>
              <Badge variant="outline">{lesson.duration}</Badge>
            </div>

            <div className="mt-4 grid gap-4">
              <LessonBlock title="学完你应该会" icon={CircleDotIcon} items={lesson.outcomes} />
              <LessonBlock
                title="配套资源"
                icon={BookMarkedIcon}
                items={lesson.resources}
                kind="badges"
              />
              <LessonBlock title="练习任务" icon={PlayCircleIcon} items={lesson.tasks} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SectionTitle({ title }: { title: string }) {
  return <div className="text-sm font-medium">{title}</div>;
}

function StageMetaCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Clock3Icon;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-xl border bg-muted/20 p-4">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon className="size-4" />
        {label}
      </div>
      <div className="mt-2 text-sm font-medium">{value}</div>
    </div>
  );
}

function LessonBlock({
  title,
  icon: Icon,
  items,
  kind = "list",
}: {
  title: string;
  icon: typeof Clock3Icon;
  items: string[];
  kind?: "list" | "badges";
}) {
  return (
    <div>
      <div className="mb-2 flex items-center gap-2 text-sm font-medium">
        <Icon className="size-4" />
        {title}
      </div>
      {kind === "badges" ? (
        <div className="flex flex-wrap gap-2">
          {items.map((item) => (
            <Badge key={item} variant="secondary">
              {item}
            </Badge>
          ))}
        </div>
      ) : (
        <div className="flex flex-col gap-2 text-sm text-muted-foreground">
          {items.map((item) => (
            <div key={item}>{item}</div>
          ))}
        </div>
      )}
    </div>
  );
}
