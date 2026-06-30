import {
  CheckCircle2Icon,
  CircleAlertIcon,
  CircleDashedIcon,
  GraduationCapIcon,
  PlayCircleIcon,
  RotateCcwIcon,
} from "lucide-react";

import type { ExerciseResult, GuidePracticePoint, LearningObjective } from "@/api/learning";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

import { masteryLabels, verdictLabels } from "../_shared";
import { FeynmanAnswerEditor } from "./feynman-answer-editor";

export function PracticePanel({
  answer,
  result,
  disabled,
  isSubmitting,
  objective,
  practicePoint,
  tutorAdvice,
  isTutorTeaching,
  onAnswerChange,
  onSubmit,
  onReset,
  onRequestTutorTeaching,
}: {
  answer: string;
  result: ExerciseResult | null;
  disabled: boolean;
  isSubmitting: boolean;
  objective: LearningObjective;
  practicePoint: GuidePracticePoint | null;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  onAnswerChange: (value: string) => void;
  onSubmit: () => void;
  onReset: () => void;
  onRequestTutorTeaching: () => void;
}) {
  return (
    <section className="flex min-h-0 flex-col overflow-hidden bg-background">
      <ScrollArea className="min-h-0 flex-1">
        <div className="flex min-h-full flex-col gap-3">
          <div className="flex shrink-0 items-start gap-2 rounded-lg bg-muted/30 px-3 py-2">
            <PlayCircleIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
            <p className="text-sm leading-6 text-muted-foreground">
              {practicePoint
                ? `本轮只复述「${practicePoint.title}」：${practicePoint.goal || "用自己的话讲清这个小点是什么、为什么重要、容易错在哪里。"}`
                : `本轮复述「${objective.title}」。如果内容太多，先回到阅读页让导学 Agent 拆成小点，再选择一个小点来讲。`}
            </p>
          </div>

          <FeynmanAnswerEditor value={answer} onChange={onAnswerChange} />

          <div className="flex shrink-0 items-center justify-between gap-3">
            <Button variant="outline" onClick={onReset} disabled={!answer && !result}>
              <RotateCcwIcon className="size-4" />
              重来
            </Button>
            <Button onClick={onSubmit} disabled={disabled || !answer.trim()}>
              {isSubmitting ? "判定中..." : "提交解释"}
            </Button>
          </div>

          {result ? (
            <div className="shrink-0">
              <ResultPanel
                result={result}
                tutorAdvice={tutorAdvice}
                isTutorTeaching={isTutorTeaching}
                onRequestTutorTeaching={onRequestTutorTeaching}
              />
            </div>
          ) : null}
        </div>
      </ScrollArea>
    </section>
  );
}

function ResultPanel({
  result,
  tutorAdvice,
  isTutorTeaching,
  onRequestTutorTeaching,
}: {
  result: ExerciseResult;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  onRequestTutorTeaching: () => void;
}) {
  const Icon =
    result.verdict === "pass"
      ? CheckCircle2Icon
      : result.verdict === "partial"
        ? CircleAlertIcon
        : CircleDashedIcon;
  const needsTeaching = result.verdict !== "pass";

  return (
    <div className="flex flex-col gap-4 rounded-xl border p-4">
      <div className="flex flex-wrap items-center gap-2">
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

      {needsTeaching ? (
        <div className="flex flex-col gap-3 rounded-lg bg-muted/30 p-3">
          <div className="flex items-center justify-between gap-3">
            <div className="text-sm font-medium">老师补讲</div>
            <Button
              variant="outline"
              size="sm"
              onClick={onRequestTutorTeaching}
              disabled={isTutorTeaching}
            >
              <GraduationCapIcon className="size-4" />
              {isTutorTeaching ? "讲解中..." : tutorAdvice ? "重新讲一下" : "让老师教我"}
            </Button>
          </div>
          {tutorAdvice || isTutorTeaching ? (
            <MessageResponse className="max-w-none text-sm leading-6 text-muted-foreground">
              {tutorAdvice || "老师正在组织讲解..."}
            </MessageResponse>
          ) : (
            <p className="text-sm leading-6 text-muted-foreground">
              这次还没讲清楚时，可以让 Tutor agent 直接补讲并给你一个小复述练习。
            </p>
          )}
        </div>
      ) : null}
    </div>
  );
}
