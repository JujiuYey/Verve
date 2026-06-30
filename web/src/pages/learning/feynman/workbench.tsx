import { useNavigate, useParams } from "@tanstack/react-router";
import {
  ArrowLeftIcon,
  BookOpenTextIcon,
  CheckCircle2Icon,
  CircleAlertIcon,
  CircleDashedIcon,
  MessageSquareTextIcon,
  PlayCircleIcon,
  RotateCcwIcon,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import {
  useCreateSession,
  useGoalDetail,
  useSubmitExercise,
  type ExerciseResult,
  type LearningObjective,
} from "@/api/learning";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";

type WorkbenchPhase = "reading" | "answering";

const masteryLabels: Record<string, string> = {
  none: "未验证",
  seen: "看过",
  heard: "听过",
  explained: "能解释",
  written: "能写出",
  verified: "已验证",
};

const verdictLabels: Record<string, string> = {
  pass: "通过",
  partial: "部分掌握",
  fail: "未通过",
};

export function FeynmanWorkbenchPage() {
  const navigate = useNavigate();
  const { goalId, objectiveId } = useParams({
    from: "/_layout/learn/feynman-practice/$goalId/$objectiveId",
  });
  const { data: detail, isLoading } = useGoalDetail(goalId);
  const createSession = useCreateSession();

  const [sessionId, setSessionId] = useState("");
  const [phase, setPhase] = useState<WorkbenchPhase>("reading");
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState<ExerciseResult | null>(null);
  const submitExercise = useSubmitExercise(sessionId);

  const objectives = useMemo(
    () => [...(detail?.objectives ?? [])].sort((a, b) => a.order_index - b.order_index),
    [detail?.objectives],
  );
  const objective = objectives.find((item) => item.id === objectiveId) ?? null;
  const stageObjectives = objectives.filter((item) => item.stage_title === objective?.stage_title);

  useEffect(() => {
    if (!objectiveId || sessionId || createSession.isPending) return;
    createSession
      .mutateAsync({ objective_id: objectiveId })
      .then((res) => setSessionId(res.session_id))
      .catch(() => toast.error("创建练习会话失败"));
  }, [createSession, objectiveId, sessionId]);

  useEffect(() => {
    setPhase("reading");
    setAnswer("");
    setResult(null);
  }, [objectiveId]);

  const submit = async () => {
    if (!sessionId || !objective || !answer.trim()) return;
    const res = await submitExercise.mutateAsync({
      type: "explain",
      prompt: buildPrompt(objective),
      user_answer: answer,
    });
    setResult(res);
  };

  const resetAnswer = () => {
    setAnswer("");
    setResult(null);
  };

  if (isLoading) {
    return (
      <div className="flex h-full flex-col gap-4 p-6">
        <Skeleton className="h-10 w-48" />
        <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(380px,0.8fr)_320px]">
          <Skeleton className="h-full rounded-2xl" />
          <Skeleton className="h-full rounded-2xl" />
          <Skeleton className="h-full rounded-2xl" />
        </div>
      </div>
    );
  }

  if (!objective) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
        <CircleAlertIcon className="size-8 text-muted-foreground" />
        <div className="text-lg font-semibold">没有找到这个小目标</div>
        <Button variant="outline" onClick={() => navigate({ to: "/learn/feynman" })}>
          返回费曼练习
        </Button>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-4 overflow-hidden p-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0">
          <Button
            variant="ghost"
            className="mb-2 px-0"
            onClick={() => navigate({ to: "/learn/feynman" })}
          >
            <ArrowLeftIcon className="size-4" />
            返回费曼练习
          </Button>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="secondary">{objective.stage_title || "学习阶段"}</Badge>
            <Badge variant="outline">
              {masteryLabels[objective.mastery_level] ?? objective.mastery_level}
            </Badge>
          </div>
          <h1 className="mt-2 truncate text-2xl font-bold">{objective.title}</h1>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <PhaseBadge phase={phase} />
          <Button
            variant="outline"
            onClick={() => navigate({ to: "/learn/goal/$goalId", params: { goalId } })}
          >
            查看路线图
          </Button>
        </div>
      </div>

      {phase === "reading" ? (
        <SourcePanel
          objective={objective}
          stageObjectives={stageObjectives}
          onStartAnswering={() => setPhase("answering")}
        />
      ) : (
        <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
          <PracticePanel
            answer={answer}
            result={result}
            disabled={!sessionId || submitExercise.isPending}
            isSubmitting={submitExercise.isPending}
            onAnswerChange={setAnswer}
            onSubmit={submit}
            onReset={resetAnswer}
            onBackToReading={() => setPhase("reading")}
          />
          <StudyInfoPanel objective={objective} result={result} sessionId={sessionId} />
        </div>
      )}
    </div>
  );
}

function PhaseBadge({ phase }: { phase: WorkbenchPhase }) {
  return (
    <div className="flex items-center gap-1 rounded-md border bg-background p-1 text-xs">
      <span
        className={`rounded px-2 py-1 ${
          phase === "reading" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
      >
        1 阅读
      </span>
      <span
        className={`rounded px-2 py-1 ${
          phase === "answering" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
      >
        2 复述
      </span>
    </div>
  );
}

function SourcePanel({
  objective,
  stageObjectives,
  onStartAnswering,
}: {
  objective: LearningObjective;
  stageObjectives: LearningObjective[];
  onStartAnswering: () => void;
}) {
  const sourceMarkdown =
    objective.detail?.trim() ||
    "当前小目标还没有展开说明。后续接入原始 Markdown 正文后，这里会展示对应资料片段。";

  return (
    <Card className="min-h-0 flex-1 rounded-2xl py-0">
      <CardHeader className="border-b p-4!">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <BookOpenTextIcon className="size-4" />
            阅读原始资料
          </CardTitle>
          <Button onClick={onStartAnswering}>
            <PlayCircleIcon className="size-4" />
            开始复述
          </Button>
        </div>
      </CardHeader>
      <CardContent className="min-h-0 p-0">
        <ScrollArea className="h-full">
          <div className="mx-auto max-w-5xl space-y-6 p-4 lg:p-6">
            <section className="rounded-xl bg-muted/40 p-4">
              <div className="mb-2 text-sm font-medium">本次小目标</div>
              <div className="text-lg font-semibold">{objective.title}</div>
            </section>

            <section className="rounded-xl border bg-background p-5">
              <MessageResponse className="max-w-none text-sm leading-7">
                {sourceMarkdown}
              </MessageResponse>
            </section>

            <section>
              <div className="mb-3 text-sm font-medium">同阶段上下文</div>
              <div className="space-y-2">
                {stageObjectives.map((item) => (
                  <div
                    key={item.id}
                    className={`rounded-lg border p-3 text-sm ${
                      item.id === objective.id ? "border-primary bg-primary/5" : "bg-background"
                    }`}
                  >
                    <div className="font-medium">{item.title}</div>
                    {item.detail ? (
                      <div className="mt-1 line-clamp-2 leading-6 text-muted-foreground">
                        {item.detail}
                      </div>
                    ) : null}
                  </div>
                ))}
              </div>
            </section>

            <div className="flex justify-end border-t pt-4">
              <Button onClick={onStartAnswering}>
                <PlayCircleIcon className="size-4" />
                我读完了，开始复述
              </Button>
            </div>
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

function PracticePanel({
  answer,
  result,
  disabled,
  isSubmitting,
  onAnswerChange,
  onSubmit,
  onReset,
  onBackToReading,
}: {
  answer: string;
  result: ExerciseResult | null;
  disabled: boolean;
  isSubmitting: boolean;
  onAnswerChange: (value: string) => void;
  onSubmit: () => void;
  onReset: () => void;
  onBackToReading: () => void;
}) {
  return (
    <Card className="min-h-0 rounded-2xl py-0">
      <CardHeader className="border-b p-4!">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <MessageSquareTextIcon className="size-4" />
            费曼复述
          </CardTitle>
          <Button variant="outline" size="sm" onClick={onBackToReading}>
            <BookOpenTextIcon className="size-4" />
            返回阅读
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex min-h-0 flex-col gap-4 p-4">
        <div className="rounded-xl border bg-muted/20 p-4">
          <div className="mb-2 flex items-center gap-2 text-sm font-medium">
            <PlayCircleIcon className="size-4" />
            本轮任务
          </div>
          <p className="text-sm leading-6 text-muted-foreground">
            假设你在教一个完全没学过的人。不要照抄原文，尽量用自己的话解释这个知识点是什么、为什么重要、容易错在哪里。
          </p>
        </div>

        <Textarea
          className="min-h-56 flex-1 resize-none"
          placeholder="把你的解释写在这里。讲不出来也可以直接写卡住的地方。"
          value={answer}
          onChange={(event) => onAnswerChange(event.target.value)}
        />

        <div className="flex items-center justify-between gap-3">
          <Button variant="outline" onClick={onReset} disabled={!answer && !result}>
            <RotateCcwIcon className="size-4" />
            重来
          </Button>
          <Button onClick={onSubmit} disabled={disabled || !answer.trim()}>
            {isSubmitting ? "判定中..." : "提交解释"}
          </Button>
        </div>

        {result ? <ResultPanel result={result} /> : null}
      </CardContent>
    </Card>
  );
}

function ResultPanel({ result }: { result: ExerciseResult }) {
  const Icon =
    result.verdict === "pass"
      ? CheckCircle2Icon
      : result.verdict === "partial"
        ? CircleAlertIcon
        : CircleDashedIcon;

  return (
    <div className="rounded-xl border p-4">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Badge
          variant={
            result.verdict === "pass"
              ? "default"
              : result.verdict === "fail"
                ? "destructive"
                : "secondary"
          }
        >
          <Icon className="size-3.5" />
          {verdictLabels[result.verdict] ?? result.verdict}
        </Badge>
        <Badge variant="outline">
          掌握度：{masteryLabels[result.mastery_after] ?? result.mastery_after}
        </Badge>
      </div>
      <p className="whitespace-pre-wrap text-sm leading-6 text-muted-foreground">
        {result.feedback}
      </p>
    </div>
  );
}

function StudyInfoPanel({
  objective,
  result,
  sessionId,
}: {
  objective: LearningObjective;
  result: ExerciseResult | null;
  sessionId: string;
}) {
  return (
    <Card className="min-h-0 rounded-2xl py-0">
      <CardHeader className="border-b p-4!">
        <CardTitle className="text-base">本次学习信息</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4 p-4">
        <InfoRow label="会话状态" value={sessionId ? "已创建" : "创建中"} />
        <InfoRow label="小目标状态" value={objective.status} />
        <InfoRow
          label="原掌握度"
          value={masteryLabels[objective.mastery_level] ?? objective.mastery_level}
        />
        <Separator />
        <InfoRow
          label="本次判定"
          value={result ? (verdictLabels[result.verdict] ?? result.verdict) : "待提交"}
        />
        <InfoRow
          label="判定后掌握度"
          value={
            result
              ? (masteryLabels[result.mastery_after] ?? result.mastery_after)
              : (masteryLabels[objective.mastery_level] ?? objective.mastery_level)
          }
        />
        <div className="rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
          通过后系统会把这次解释写入练习记录。后续接入 Learning Examiner
          后，这里会同步展示薄弱点、日志和下一步建议。
        </div>
      </CardContent>
    </Card>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="max-w-44 truncate font-medium">{value}</span>
    </div>
  );
}

function buildPrompt(objective: LearningObjective) {
  return `请用自己的话解释：${objective.title}`;
}
