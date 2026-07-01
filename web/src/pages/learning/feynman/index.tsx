import { useNavigate } from "@tanstack/react-router";
import {
  ArrowRightIcon,
  BrainCircuitIcon,
  CheckCircle2Icon,
  CircleAlertIcon,
  CircleDashedIcon,
  HistoryIcon,
  Loader2Icon,
  PlayCircleIcon,
  RouteIcon,
  SparklesIcon,
} from "lucide-react";
import { useState } from "react";

import {
  useCreateGoal,
  useLearningOrchestrator,
  useOrchestrateLearning,
  type LearningOrchestratorAction,
  type LearningExercise,
} from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";

const verdictMeta = {
  pass: { label: "通过", icon: CheckCircle2Icon, className: "text-foreground" },
  partial: { label: "部分掌握", icon: CircleAlertIcon, className: "text-foreground" },
  fail: { label: "未通过", icon: CircleDashedIcon, className: "text-muted-foreground" },
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
  const navigate = useNavigate();
  const [learningIntent, setLearningIntent] = useState("");
  const [orchestratedActions, setOrchestratedActions] = useState<LearningOrchestratorAction[] | null>(
    null,
  );
  const { data: orchestratorData, isLoading: orchestratorLoading } = useLearningOrchestrator();
  const orchestrateLearning = useOrchestrateLearning();
  const createGoal = useCreateGoal();

  const actions = orchestratedActions ?? orchestratorData?.actions ?? [];
  const recentExercises = orchestratorData?.recent ?? [];
  const canSubmitIntent = learningIntent.trim().length > 0 && !orchestrateLearning.isPending;
  const isActionPending = createGoal.isPending || orchestrateLearning.isPending;

  const startFromIntent = () => {
    const intent = learningIntent.trim();
    if (!intent) return;

    orchestrateLearning.mutate(
      { intent },
      {
        onSuccess: (result) => {
          setOrchestratedActions(result.actions);
        },
      },
    );
  };

  const runAction = (action: LearningOrchestratorAction) => {
    if (action.type === "create_goal") {
      const title = (action.intent || learningIntent).trim();
      if (!title) return;

      createGoal.mutate(
        { title },
        {
          onSuccess: ({ goal_id }) => {
            navigate({ to: "/learn/goal/$goalId", params: { goalId: goal_id } });
          },
        },
      );
      return;
    }

    if (
      (action.type === "continue_objective" || action.type === "review_objective") &&
      action.goal_id &&
      action.objective_id
    ) {
      navigate({
        to: "/learn/feynman-practice/$goalId/$objectiveId",
        params: { goalId: action.goal_id, objectiveId: action.objective_id },
      });
      return;
    }

    if (action.goal_id) {
      navigate({ to: "/learn/goal/$goalId", params: { goalId: action.goal_id } });
    }
  };

  return (
    <div className="flex h-full flex-col gap-6 overflow-auto p-6">
      <div className="flex flex-col gap-2 px-1">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <BrainCircuitIcon className="size-4" />
          学习调度
        </div>
        <h1 className="text-2xl font-bold">今天学点什么</h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
          这里是你的费曼学习入口。先说想学什么，系统会把它变成学习路线；如果已经有进度，就直接选择下一步进入练习。
        </p>
      </div>

      <div className="grid min-h-0 flex-1 gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="flex min-h-0 flex-col gap-5">
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <SparklesIcon className="size-4" />
                学习调度 agent
              </CardTitle>
              <CardDescription>
                输入一个真实意图，调度器会结合学习画像、最近验证和当前路线给出下一步。
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <Textarea
                value={learningIntent}
                onChange={(event) => {
                  setLearningIntent(event.target.value);
                  setOrchestratedActions(null);
                }}
                placeholder="今天学点什么？比如：我想把 Go 的接口、context 和错误处理真正讲明白"
                className="min-h-36 resize-none text-sm leading-6"
                onKeyDown={(event) => {
                  if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
                    startFromIntent();
                  }
                }}
              />
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="text-xs text-muted-foreground">
                  {orchestrateLearning.data?.summary ??
                    orchestratorData?.summary ??
                    "具体解释、判定和补讲会在练习工作台里完成。"}
                </div>
                <Button disabled={!canSubmitIntent} onClick={startFromIntent}>
                  {orchestrateLearning.isPending ? (
                    <Loader2Icon data-icon="inline-start" className="animate-spin" />
                  ) : (
                    <RouteIcon data-icon="inline-start" />
                  )}
                  让 agent 调度
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card className="min-h-0 rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">继续选项</CardTitle>
              <CardDescription>
                {orchestrateLearning.data?.habit_summary ??
                  orchestratorData?.habit_summary ??
                  "正在读取你的学习习惯和最近验证。"}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {orchestratorLoading ? (
                <div className="flex flex-col gap-3">
                  {Array.from({ length: 3 }).map((_, index) => (
                    <Skeleton key={index} className="h-20 w-full" />
                  ))}
                </div>
              ) : (
                <>
                  {actions.map((action) => (
                    <button
                      key={action.id}
                      type="button"
                      disabled={isActionPending}
                      className="rounded-xl border p-4 text-left transition-colors hover:border-primary hover:bg-primary/5 focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-60"
                      onClick={() => runAction(action)}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="mb-2 flex items-center gap-2">
                            <Badge>{action.label}</Badge>
                            <span className="text-xs text-muted-foreground">{action.reason}</span>
                          </div>
                          <div className="font-medium">{action.title}</div>
                          <div className="mt-1 text-sm leading-6 text-muted-foreground">
                            {action.description}
                          </div>
                        </div>
                        {isActionPending && action.type === "create_goal" ? (
                          <Loader2Icon className="mt-1 size-4 shrink-0 animate-spin text-muted-foreground" />
                        ) : (
                          <ArrowRightIcon className="mt-1 size-4 shrink-0 text-muted-foreground" />
                        )}
                      </div>
                    </button>
                  ))}

                  {actions.length === 0 ? (
                    <Empty className="min-h-48 border">
                      <EmptyHeader>
                        <EmptyMedia variant="icon">
                          <PlayCircleIcon />
                        </EmptyMedia>
                        <EmptyTitle>还没有可继续的学习路线</EmptyTitle>
                        <EmptyDescription>
                          在上面的输入框写下今天想学什么，先生成第一条路线。
                        </EmptyDescription>
                      </EmptyHeader>
                    </Empty>
                  ) : null}
                </>
              )}
            </CardContent>
          </Card>
        </div>

        <Card className="flex min-h-0 flex-col overflow-hidden rounded-2xl">
          <CardHeader className="shrink-0">
            <CardTitle className="flex items-center gap-2 text-base">
              <HistoryIcon className="size-4" />
              最近验证
            </CardTitle>
            <CardDescription>这里不是主菜单，只用来回看最近几次解释判定。</CardDescription>
          </CardHeader>
          <CardContent className="min-h-0 flex-1">
            {orchestratorLoading ? (
              <ScrollArea className="h-full">
                <div className="flex flex-col gap-3 pr-3">
                  {Array.from({ length: 4 }).map((_, index) => (
                    <Skeleton key={index} className="h-24 w-full" />
                  ))}
                </div>
              </ScrollArea>
            ) : recentExercises.length === 0 ? (
              <Empty className="min-h-56 border">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <HistoryIcon />
                  </EmptyMedia>
                  <EmptyTitle>还没有验证记录</EmptyTitle>
                  <EmptyDescription>完成一次费曼复述后，最近结果会出现在这里。</EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : (
              <ScrollArea className="h-full">
                <div className="flex flex-col gap-3 pr-3">
                  {recentExercises.map((exercise) => (
                    <ExerciseRecord key={exercise.id} exercise={exercise} />
                  ))}
                </div>
              </ScrollArea>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function ExerciseRecord({ exercise }: { exercise: LearningExercise }) {
  const meta = exercise.verdict ? verdictMeta[exercise.verdict as keyof typeof verdictMeta] : null;
  const Icon = meta?.icon ?? CircleDashedIcon;

  return (
    <div className="rounded-xl border p-4">
      <div className="flex items-start justify-between gap-3">
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
        </div>
        <div className="flex shrink-0 items-center gap-2 rounded-lg bg-muted/40 px-2 py-1.5">
          <Icon className={`size-4 ${meta?.className ?? "text-muted-foreground"}`} />
          <span className="text-xs font-medium">{meta?.label ?? "待判定"}</span>
        </div>
      </div>
      {exercise.feedback || exercise.mastery_after ? (
        <>
          <Separator className="my-3" />
          <div className="flex flex-col gap-1">
            {exercise.feedback ? (
              <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">
                {exercise.feedback}
              </p>
            ) : null}
            <div className="text-xs text-muted-foreground">
              掌握状态：
              {exercise.mastery_after
                ? (masteryLabels[exercise.mastery_after] ?? exercise.mastery_after)
                : "未验证"}
            </div>
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
