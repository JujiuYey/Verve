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

import type { ExerciseResult, GuidePracticePoint, LearningObjective } from "@/api/learning";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

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
  canAppendTutorNote,
  isAppendingTutorNote,
  onAnswerChange,
  onSubmit,
  onReset,
  onRequestTutorTeaching,
  onAppendTutorNote,
}: {
  answer: string;
  result: ExerciseResult | null;
  disabled: boolean;
  isSubmitting: boolean;
  objective: LearningObjective;
  practicePoint: GuidePracticePoint | null;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onAnswerChange: (value: string) => void;
  onSubmit: () => void;
  onReset: () => void;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
}) {
  return (
    <section className="flex min-h-0 flex-col overflow-hidden bg-background">
      <ScrollArea className="min-h-0 flex-1 px-4">
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
}: {
  result: ExerciseResult;
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
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
        <LearningLoopTabs
          tutorAdvice={tutorAdvice}
          isTutorTeaching={isTutorTeaching}
          canAppendTutorNote={canAppendTutorNote}
          isAppendingTutorNote={isAppendingTutorNote}
          onRequestTutorTeaching={onRequestTutorTeaching}
          onAppendTutorNote={onAppendTutorNote}
        />
      ) : null}
    </div>
  );
}

function LearningLoopTabs({
  tutorAdvice,
  isTutorTeaching,
  canAppendTutorNote,
  isAppendingTutorNote,
  onRequestTutorTeaching,
  onAppendTutorNote,
}: {
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
}) {
  return (
    <Tabs defaultValue="read" className="rounded-lg bg-muted/30 p-3">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <TabsList className="grid w-full grid-cols-3 sm:w-fit">
          <TabsTrigger value="read">
            <NotebookPenIcon />
            阅读
          </TabsTrigger>
          <TabsTrigger value="retell">
            <Repeat2Icon />
            复述
          </TabsTrigger>
          <TabsTrigger value="teach">
            <GraduationCapIcon />
            讲解
          </TabsTrigger>
        </TabsList>
        <Button
          variant="outline"
          size="sm"
          onClick={onRequestTutorTeaching}
          disabled={isTutorTeaching}
        >
          <GraduationCapIcon data-icon="inline-start" />
          {isTutorTeaching ? "讲解中..." : tutorAdvice ? "重新讲一下" : "让老师讲解"}
        </Button>
      </div>

      <TabsContent value="read" className="mt-3">
        <LoopNote
          title="回到教材"
          description="先把原文里没读透的定义、例子、边界条件重新读一遍。这里的目标不是刷进度，而是找出自己复述时卡住的那一句。"
        />
      </TabsContent>
      <TabsContent value="retell" className="mt-3">
        <LoopNote
          title="重新用自己的话讲"
          description="把刚才漏掉的关键点补进解释里，再提交一次。能讲出是什么、为什么、怎么用、哪里容易错，就说明这一轮可以往前走。"
        />
      </TabsContent>
      <TabsContent value="teach" className="mt-3">
        <div className="flex flex-col gap-3">
          {tutorAdvice || isTutorTeaching ? (
            <MessageResponse className="max-w-none text-sm leading-6 text-muted-foreground">
              {tutorAdvice || "老师正在组织讲解..."}
            </MessageResponse>
          ) : (
            <LoopNote
              title="让老师补一段可沉淀的讲解"
              description="讲解会按教材旁注的方式组织：先补清楚知识点，再指出你漏掉的地方，最后给出可以写回 Markdown 的用户笔记。"
            />
          )}
          <div className="rounded-md border bg-background px-3 py-2 text-xs leading-5 text-muted-foreground">
            这部分更像用户写在教材边上的笔记：不是替换原教材，而是把真实学习时踩过的坑、补上的例子和更顺手的解释沉淀回
            Markdown。
          </div>
          <Button
            variant="secondary"
            size="sm"
            className="self-start"
            onClick={onAppendTutorNote}
            disabled={!canAppendTutorNote || !tutorAdvice.trim() || isAppendingTutorNote}
          >
            <NotebookPenIcon data-icon="inline-start" />
            {isAppendingTutorNote ? "追加中..." : "追加到 Markdown"}
          </Button>
        </div>
      </TabsContent>
    </Tabs>
  );
}

function LoopNote({ title, description }: { title: string; description: string }) {
  return (
    <div className="rounded-md border bg-background px-3 py-2">
      <div className="text-sm font-medium">{title}</div>
      <p className="mt-1 text-sm leading-6 text-muted-foreground">{description}</p>
    </div>
  );
}
