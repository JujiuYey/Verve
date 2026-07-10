import { CircleAlertIcon, ListChecksIcon, RouteIcon, TargetIcon, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

import type { LearningObjective } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import { masteryLabels } from "../_shared";

const watchedMasteryLevels = new Set(["seen", "heard", "explained", "written", "verified"]);
const practicedMasteryLevels = new Set(["heard", "explained", "written", "verified"]);

export function ObjectiveOutline({
  objective,
  objectives,
  isLoading,
  previousObjective,
  nextObjective,
  onOpenObjective,
}: {
  objective: LearningObjective;
  objectives: LearningObjective[];
  isLoading: boolean;
  previousObjective: LearningObjective | null;
  nextObjective: LearningObjective | null;
  onOpenObjective: (id: string) => void;
}) {
  const items = objectives.length > 0 ? objectives : [objective];

  return (
    <div className="flex flex-col gap-4 p-4">
      <Section icon={TargetIcon} title="当前要掌握">
        <div className="rounded-md border border-primary/30 bg-primary/5 p-3">
          <div className="text-sm font-medium leading-5">{objective.title}</div>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {objective.detail || "用自己的话讲清这个小节的定义、用途、边界和易错点。"}
          </p>
          <Badge variant="outline" className="mt-3">
            {masteryLabels[objective.mastery_level] ?? objective.mastery_level}
          </Badge>
        </div>
      </Section>

      <Section icon={ListChecksIcon} title="文档小节">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {items.map((item, index) => {
              const active = item.id === objective.id;
              return (
                <button
                  key={item.id}
                  type="button"
                  className={cn(
                    "w-full rounded-md border bg-background px-3 py-2 text-left transition-colors hover:border-primary/60 hover:bg-primary/5",
                    active && "border-primary bg-primary/10",
                  )}
                  onClick={() => onOpenObjective(item.id)}
                >
                  <div className="flex items-start gap-2">
                    <span
                      className={cn(
                        "mt-1 flex size-5 shrink-0 items-center justify-center rounded-full border text-xs font-medium",
                        active
                          ? "border-primary bg-primary text-primary-foreground"
                          : "bg-muted text-muted-foreground",
                      )}
                    >
                      {index + 1}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="flex min-w-0 flex-wrap items-center gap-1.5">
                        <span className="min-w-0 truncate text-sm font-medium leading-5">
                          {item.title}
                        </span>
                        <ObjectiveProgressBadges objective={item} />
                      </span>
                      {item.detail ? (
                        <span className="mt-1 line-clamp-3 block text-xs leading-5 text-muted-foreground">
                          {item.detail}
                        </span>
                      ) : null}
                    </span>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </Section>

      <Section icon={RouteIcon} title="阶段位置">
        <div className="flex flex-col gap-2 text-sm">
          <ContextRow label="上一小节" value={previousObjective?.title || "这是当前文档开头"} />
          <ContextRow label="当前小节" value={objective.title} active />
          <ContextRow label="下一小节" value={nextObjective?.title || "读完后进入复述验证"} />
        </div>
      </Section>

      <div className="rounded-md border bg-background p-3">
        <div className="mb-2 flex items-center gap-2 text-sm font-medium">
          <CircleAlertIcon className="size-4 text-primary" />
          复述前自检
        </div>
        <p className="text-sm leading-6 text-muted-foreground">
          先只围绕当前小节复述。能讲出是什么、为什么、怎么用、哪里容易错，再进入复述验证。
        </p>
      </div>
    </div>
  );
}

function ObjectiveProgressBadges({ objective }: { objective: LearningObjective }) {
  const badges: string[] = [];

  if (watchedMasteryLevels.has(objective.mastery_level)) {
    badges.push("看过");
  }
  if (practicedMasteryLevels.has(objective.mastery_level) || objective.status === "completed") {
    badges.push("练过");
  }

  if (badges.length === 0) return null;

  return (
    <span className="inline-flex shrink-0 flex-wrap items-center gap-1">
      {badges.map((label) => (
        <Badge
          key={label}
          variant="secondary"
          className="h-5 rounded-full px-1.5 text-[11px] font-normal leading-none"
        >
          {label}
        </Badge>
      ))}
    </span>
  );
}

function Section({
  icon: Icon,
  title,
  children,
}: {
  icon: LucideIcon;
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="flex flex-col gap-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Icon className="size-4 text-primary" />
        {title}
      </div>
      {children}
    </section>
  );
}

function ContextRow({ label, value, active }: { label: string; value: string; active?: boolean }) {
  return (
    <div className={`rounded-md border p-2 ${active ? "border-primary bg-primary/5" : ""}`}>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 line-clamp-2 font-medium leading-5">{value}</div>
    </div>
  );
}
