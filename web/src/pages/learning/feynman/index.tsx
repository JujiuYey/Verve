import {
  ArrowRightIcon,
  CheckCircle2Icon,
  CircleAlertIcon,
  CircleDashedIcon,
  ListChecksIcon,
  PlayCircleIcon,
} from "lucide-react";

import { useExerciseList, type LearningExercise } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

const verdictMeta = {
  pass: { label: "通过", icon: CheckCircle2Icon, className: "text-emerald-600" },
  partial: { label: "部分掌握", icon: CircleAlertIcon, className: "text-amber-600" },
  fail: { label: "未通过", icon: CircleDashedIcon, className: "text-destructive" },
};

const masteryLabels: Record<string, string> = {
  none: "未验证",
  seen: "看过",
  heard: "听过",
  explained: "能解释",
  written: "能写出",
  verified: "已验证",
};

export function FeynmanExercisePage() {
  const { data: exercisesData, isLoading: exercisesLoading } = useExerciseList(1, 30);

  const exercises = exercisesData?.data ?? [];

  return (
    <div className="flex h-full flex-col gap-6 overflow-auto p-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <div className="mb-2 flex items-center gap-2 text-sm text-muted-foreground">
            <ListChecksIcon className="size-4" />
            费曼练习
          </div>
          <h1 className="text-2xl font-bold">用自己的话讲明白</h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
            这里记录每次解释验证的结果。新的练习从学习项目详情里的小目标进入。
          </p>
        </div>
      </div>

      <Card className="min-h-0 flex-1 rounded-2xl">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">练习记录</CardTitle>
          <Badge variant="secondary">{exercisesData?.total ?? 0} 条</Badge>
        </CardHeader>
        <CardContent>
          {exercisesLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, index) => (
                <Skeleton key={index} className="h-24 w-full" />
              ))}
            </div>
          ) : exercises.length === 0 ? (
            <div className="flex min-h-48 flex-col items-center justify-center rounded-xl border border-dashed text-center">
              <PlayCircleIcon className="mb-3 size-8 text-muted-foreground" />
              <div className="font-medium">还没有费曼练习记录</div>
              <div className="mt-1 text-sm text-muted-foreground">
                从学习项目详情选择一个小目标，完成第一次解释验证后会出现在这里。
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {exercises.map((exercise) => (
                <ExerciseRecord key={exercise.id} exercise={exercise} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function ExerciseRecord({ exercise }: { exercise: LearningExercise }) {
  const meta = exercise.verdict ? verdictMeta[exercise.verdict as keyof typeof verdictMeta] : null;
  const Icon = meta?.icon ?? CircleDashedIcon;

  return (
    <div className="rounded-xl border p-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0 flex-1">
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <Badge variant="outline">{typeLabel(exercise.type)}</Badge>
            <span className="text-xs text-muted-foreground">
              {new Date(exercise.created_at).toLocaleString()}
            </span>
          </div>
          <div className="line-clamp-2 text-sm font-medium">
            {exercise.prompt || "费曼解释验证"}
          </div>
          {exercise.user_answer ? (
            <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted-foreground">
              {exercise.user_answer}
            </p>
          ) : null}
        </div>

        <div className="flex shrink-0 items-center gap-3 rounded-lg bg-muted/30 px-3 py-2">
          <Icon className={`size-4 ${meta?.className ?? "text-muted-foreground"}`} />
          <div>
            <div className="text-sm font-medium">{meta?.label ?? "待判定"}</div>
            <div className="text-xs text-muted-foreground">
              {exercise.mastery_after
                ? (masteryLabels[exercise.mastery_after] ?? exercise.mastery_after)
                : "未验证"}
            </div>
          </div>
        </div>
      </div>
      {exercise.feedback ? (
        <>
          <Separator className="my-3" />
          <div className="flex items-start justify-between gap-3">
            <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">
              {exercise.feedback}
            </p>
            <ArrowRightIcon className="mt-1 size-4 shrink-0 text-muted-foreground" />
          </div>
        </>
      ) : null}
    </div>
  );
}

function typeLabel(type: string) {
  const labels: Record<string, string> = {
    explain: "解释",
    choice: "选择",
    cloze: "填空",
    paste_output: "运行结果",
    code_snippet: "代码",
  };
  return labels[type] ?? type;
}
