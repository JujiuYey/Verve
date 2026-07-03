import {
  CheckCircle2Icon,
  CircleAlertIcon,
  CircleDashedIcon,
  GraduationCapIcon,
  NotebookPenIcon,
  PlayCircleIcon,
  Repeat2Icon,
  RotateCcwIcon,
} from "lucide-react";

import type { ExerciseResult, LearningObjective } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

import { masteryLabels, verdictLabels, type WorkbenchPhase } from "../_shared";
import { FeynmanAnswerEditor } from "./feynman-answer-editor";

export function PracticePanel({
  answer,
  result,
  disabled,
  isSubmitting,
  objective,
  tutorAdvice,
  isTutorTeaching,
  canAppendTutorNote,
  isAppendingTutorNote,
  onAnswerChange,
  onSubmit,
  onReset,
  onRequestTutorTeaching,
  onAppendTutorNote,
  onPhaseChange,
}: {
  answer: string;
  result: ExerciseResult | null;
  disabled: boolean;
  isSubmitting: boolean;
  objective: LearningObjective;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onAnswerChange: (value: string) => void;
  onSubmit: () => void;
  onReset: () => void;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
  onPhaseChange: (phase: WorkbenchPhase) => void;
}) {
  return (
    <section className="flex min-h-0 flex-col overflow-hidden bg-background">
      <ScrollArea className="min-h-0 flex-1 px-4">
        <div className="flex min-h-full flex-col gap-3">
          <div className="flex shrink-0 items-start gap-2 rounded-lg bg-muted/30 px-3 py-2">
            <PlayCircleIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
            <p className="text-sm leading-6 text-muted-foreground">
              本轮只复述「{objective.title}」：
              {objective.detail ||
                "用自己的话讲清这个小节是什么、为什么重要、怎么用、容易错在哪里。"}
            </p>
          </div>

          <FeynmanAnswerEditor value={answer} onChange={onAnswerChange} />

          <div className="flex shrink-0 items-center justify-between gap-3">
            <Button variant="outline" onClick={onReset} disabled={!answer && !result}>
              <RotateCcwIcon data-icon="inline-start" />
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
                canAppendTutorNote={canAppendTutorNote}
                isAppendingTutorNote={isAppendingTutorNote}
                onRequestTutorTeaching={onRequestTutorTeaching}
                onAppendTutorNote={onAppendTutorNote}
                onPhaseChange={onPhaseChange}
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
  canAppendTutorNote,
  isAppendingTutorNote,
  onRequestTutorTeaching,
  onAppendTutorNote,
  onPhaseChange,
}: {
  result: ExerciseResult;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
  onPhaseChange: (phase: WorkbenchPhase) => void;
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
        <RecoveryTask
          tutorAdvice={tutorAdvice}
          isTutorTeaching={isTutorTeaching}
          canAppendTutorNote={canAppendTutorNote}
          isAppendingTutorNote={isAppendingTutorNote}
          onRequestTutorTeaching={onRequestTutorTeaching}
          onAppendTutorNote={onAppendTutorNote}
          onPhaseChange={onPhaseChange}
        />
      ) : null}
    </div>
  );
}

function RecoveryTask({
  tutorAdvice,
  isTutorTeaching,
  canAppendTutorNote,
  isAppendingTutorNote,
  onRequestTutorTeaching,
  onAppendTutorNote,
  onPhaseChange,
}: {
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
  onPhaseChange: (phase: WorkbenchPhase) => void;
}) {
  return (
    <div className="rounded-lg bg-muted/30 p-3">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="text-sm font-medium">重新用自己的话讲</div>
        <Button
          variant="outline"
          size="sm"
          className="shrink-0"
          onClick={() => onPhaseChange("teaching")}
        >
          <GraduationCapIcon data-icon="inline-start" />
          去教学
        </Button>
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        <Button variant="secondary" size="sm" onClick={() => onPhaseChange("reading")}>
          <NotebookPenIcon data-icon="inline-start" />
          回到阅读
        </Button>
        <Button variant="secondary" size="sm" onClick={() => onPhaseChange("answering")}>
          <Repeat2Icon data-icon="inline-start" />
          重新复述
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => {
            onPhaseChange("teaching");
            onRequestTutorTeaching();
          }}
          disabled={isTutorTeaching}
        >
          <GraduationCapIcon data-icon="inline-start" />
          {isTutorTeaching ? "讲解中..." : tutorAdvice ? "重新讲一下" : "让老师讲解"}
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={onAppendTutorNote}
          disabled={!canAppendTutorNote || !tutorAdvice.trim() || isAppendingTutorNote}
        >
          <NotebookPenIcon data-icon="inline-start" />
          {isAppendingTutorNote ? "追加中..." : "追加到 Markdown"}
        </Button>
      </div>
    </div>
  );
}