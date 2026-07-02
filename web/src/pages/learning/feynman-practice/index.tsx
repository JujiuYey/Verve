import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { ArrowLeftIcon, CircleAlertIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import {
  sessionChatStream,
  useCreateSession,
  useObjectiveDetail,
  useSubmitExercise,
  type ExerciseResult,
  type GuidePracticePoint,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

import { PhaseBadge } from "./_components/phase-badge";
import { PracticePanel } from "./_components/practice-panel";
import { SourcePanel } from "./_components/source-panel";
import { StudyInfoPanel } from "./_components/study-info-panel";
import { buildPrompt, masteryLabels, type WorkbenchPhase } from "./_shared";

export function FeynmanWorkbenchPage() {
  const navigate = useNavigate();
  const { objectiveId } = useParams({
    from: "/_layout/learn/feynman-practice/$objectiveId",
  });
  const { data: objective, isLoading } = useObjectiveDetail(objectiveId);
  const createSession = useCreateSession();
  const queryClient = useQueryClient();

  const [sessionId, setSessionId] = useState("");
  const [phase, setPhase] = useState<WorkbenchPhase>("reading");
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState<ExerciseResult | null>(null);
  const [tutorAdvice, setTutorAdvice] = useState("");
  const [isTutorTeaching, setIsTutorTeaching] = useState(false);
  const [isAppendingTutorNote, setIsAppendingTutorNote] = useState(false);
  const [selectedPracticePoint, setSelectedPracticePoint] = useState<GuidePracticePoint | null>(
    null,
  );
  const submitExercise = useSubmitExercise(sessionId);

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
    setTutorAdvice("");
    setIsTutorTeaching(false);
    setIsAppendingTutorNote(false);
    setSelectedPracticePoint(null);
  }, [objectiveId]);

  const submit = async () => {
    if (!sessionId || !objective || !answer.trim()) return;
    const res = await submitExercise.mutateAsync({
      type: "explain",
      prompt: buildPrompt(objective, selectedPracticePoint),
      user_answer: answer,
    });
    setResult(res);
    setTutorAdvice("");
    setIsTutorTeaching(false);
    void queryClient.invalidateQueries({ queryKey: ["learning-journals"] });
    void queryClient.invalidateQueries({ queryKey: ["learning-profiles"] });
  };

  const resetAnswer = () => {
    setAnswer("");
    setResult(null);
    setTutorAdvice("");
    setIsTutorTeaching(false);
    setIsAppendingTutorNote(false);
  };

  const requestTutorTeaching = async () => {
    if (!sessionId || !objective || !result || isTutorTeaching) return;

    const message = [
      `我刚才复述「${selectedPracticePoint?.title || objective.title}」没有通过。`,
      selectedPracticePoint?.goal ? `本轮目标是：${selectedPracticePoint.goal}` : "",
      `Examiner 的反馈是：${result.feedback}`,
      "请你不要只评价我，直接教我：先用通俗的话讲清楚这个知识点，再指出我漏掉的关键点，最后给我一个很小的复述练习。",
      "最后请补一段「可写回 Markdown 的学习旁注」：像用户写在教材边上的笔记一样，补充更顺手的解释、例子、易错点和复述提示，不要替换原教材。",
    ]
      .filter(Boolean)
      .join("\n");

    setTutorAdvice("");
    setIsTutorTeaching(true);
    await sessionChatStream(
      sessionId,
      message,
      (event) => {
        if ((event.type === "stream_chunk" || event.type === "message") && event.content) {
          setTutorAdvice((prev) => prev + event.content);
        } else if (event.type === "error") {
          toast.error(event.content || "教学 agent 出错");
        }
      },
      () => setIsTutorTeaching(false),
      (err) => {
        setIsTutorTeaching(false);
        toast.error(err.message);
      },
    );
  };

  const appendTutorNoteToMarkdown = async () => {
    if (!objective?.source_document_id || !tutorAdvice.trim() || isAppendingTutorNote) return;

    setIsAppendingTutorNote(true);
    try {
      const sourceDocument = await documentApi.getContent(objective.source_document_id);
      const title = selectedPracticePoint?.title || objective.title;
      const note = ["", "---", "", `## 学习旁注：${title}`, "", tutorAdvice.trim(), ""].join("\n");
      const nextContent = `${sourceDocument.content?.trimEnd() || ""}${note}`;

      await documentApi.updateContent(objective.source_document_id, { content: nextContent });
      await queryClient.invalidateQueries({
        queryKey: ["feynman-source-document", objective.source_document_id],
      });
      toast.success("学习旁注已追加到 Markdown");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "追加学习旁注失败");
    } finally {
      setIsAppendingTutorNote(false);
    }
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
          <PhaseBadge phase={phase} onPhaseChange={setPhase} />
          <Button variant="outline" onClick={() => navigate({ to: "/wiki/folders" })}>
            返回 Wiki
          </Button>
        </div>
      </div>

      {phase === "reading" ? (
        <SourcePanel
          objective={objective}
          stageObjectives={[]}
          selectedPracticePoint={selectedPracticePoint}
          onPracticePointSelect={setSelectedPracticePoint}
        />
      ) : (
        <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
          <PracticePanel
            answer={answer}
            result={result}
            disabled={!sessionId || submitExercise.isPending}
            isSubmitting={submitExercise.isPending}
            objective={objective}
            practicePoint={selectedPracticePoint}
            onAnswerChange={setAnswer}
            onSubmit={submit}
            onReset={resetAnswer}
            tutorAdvice={tutorAdvice}
            isTutorTeaching={isTutorTeaching}
            onRequestTutorTeaching={requestTutorTeaching}
            canAppendTutorNote={!!objective.source_document_id}
            isAppendingTutorNote={isAppendingTutorNote}
            onAppendTutorNote={appendTutorNoteToMarkdown}
          />
          <StudyInfoPanel objective={objective} result={result} sessionId={sessionId} />
        </div>
      )}
    </div>
  );
}
