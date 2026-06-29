import { BookOpenIcon, BrainCircuitIcon, CompassIcon, WorkflowIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";

type HeaderProps = {
  roadmapCount: number;
  stageCount: number;
};

export function LearningOverviewHeader({ roadmapCount, stageCount }: HeaderProps) {
  return (
    <section className="rounded-2xl border bg-gradient-to-br from-background via-background to-muted/50 p-6">
      <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div className="max-w-2xl space-y-3">
          <div className="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-sm text-muted-foreground">
            <CompassIcon className="size-4" />
            学习项目
          </div>
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight">做成一张能点开的学习地图</h1>
            <p className="text-sm leading-6 text-muted-foreground">
              先从项目卡片挑一个方向，再进入路线图详情页。每个项目都先用前端 mock
              数据驱动，重点把“卡片入口 + 思维导图式地图 + 节点说明”的体验搭起来。
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">
              <BookOpenIcon />
              前端实战
            </Badge>
            <Badge variant="secondary">
              <WorkflowIcon />
              工程能力
            </Badge>
            <Badge variant="secondary">
              <BrainCircuitIcon />
              AI 产品
            </Badge>
          </div>
        </div>
        <div className="grid gap-3 sm:grid-cols-3">
          <MetricCard label="项目数" value={`${roadmapCount}`} hint="先挑方向再深入" />
          <MetricCard label="路线阶段" value={`${stageCount}`} hint="每条路线分阶段推进" />
          <MetricCard label="当前形态" value="Mock" hint="后面可直接换真实接口" />
        </div>
      </div>
    </section>
  );
}

function MetricCard({ label, value, hint }: { label: string; value: string; hint: string }) {
  return (
    <div className="rounded-xl border bg-background px-4 py-3">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1 text-2xl font-semibold">{value}</div>
      <div className="mt-1 text-xs text-muted-foreground">{hint}</div>
    </div>
  );
}
