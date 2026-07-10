import { useQueryClient } from "@tanstack/react-query";
import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { CircleAlertIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import {
  exerciseKeys,
  sessionChatStream,
  objectiveKeys,
  useCreateSession,
  useObjectives,
  useSubmitExercise,
  type ExerciseResult,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

import { PhaseBadge } from "./_components/phase-badge";
import { PracticePanel } from "./_components/practice-panel";
import { SourcePanel } from "./_components/source-panel";
import { StudyInfoPanel } from "./_components/study-info-panel";
import { TeachingPanel } from "./_components/teaching-panel";
import { buildPrompt, masteryLabels, type WorkbenchPhase } from "./_shared";

const routeApi = getRouteApi("/_layout/learn/feynman-practice/$documentId");

export function FeynmanWorkbenchPage() {
  const navigate = useNavigate();
  const { documentId } = routeApi.useParams();
  const { objectiveId } = routeApi.useSearch();
  const { data: objectives = [], isLoading } = useObjectives({ document_id: documentId });
  const objective = objectives.find((item) => item.id === objectiveId) || objectives[0];
  const selectedObjectiveId = objective?.id || "";
  const createSession = useCreateSession();
  const queryClient = useQueryClient();

  const [sessionId, setSessionId] = useState("");
  const [phase, setPhase] = useState<WorkbenchPhase>("reading");
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState<ExerciseResult | null>(null);
  const [tutorAdvice, setTutorAdvice] = useState("");
  const [isTutorTeaching, setIsTutorTeaching] = useState(false);
  const [isAppendingTutorNote, setIsAppendingTutorNote] = useState(false);
  const submitExercise = useSubmitExercise(sessionId);

  useEffect(() => {
    if (!selectedObjectiveId || sessionId || createSession.isPending) return;
    createSession
      .mutateAsync({ objective_id: selectedObjectiveId })
      .then((res) => setSessionId(res.session_id))
      .catch(() => toast.error("创建练习会话失败"));
  }, [createSession, selectedObjectiveId, sessionId]);

  useEffect(() => {
    setSessionId("");
    setPhase("reading");
    setAnswer("");
    setResult(null);
    setTutorAdvice("");
    setIsTutorTeaching(false);
    setIsAppendingTutorNote(false);
  }, [selectedObjectiveId]);

  const openObjective = (id: string) => {
    if (id === selectedObjectiveId) return;
    navigate({
      to: "/learn/feynman-practice/$documentId",
      params: { documentId },
      search: { objectiveId: id },
      replace: true,
    });
  };

  const submit = async () => {
    if (!sessionId || !objective || !answer.trim()) return;
    const res = await submitExercise.mutateAsync({
      type: "explain",
      prompt: buildPrompt(objective),
      user_answer: answer,
    });
    setResult(res);
    setTutorAdvice("");
    setIsTutorTeaching(false);
    void queryClient.invalidateQueries({
      queryKey: objectiveKeys.list({ document_id: documentId }),
    });
    void queryClient.invalidateQueries({ queryKey: exerciseKeys.lists() });
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
      `我刚才复述「${objective.title}」没有通过。`,
      objective.detail ? `本轮目标是：${objective.detail}` : "",
      `Examiner 的反馈是：${result.feedback}`,
      "请你不要只评价我，直接教我：先用通俗的话讲清楚这个知识点，再指出我漏掉的关键点，最后给我一个很小的复述练习。",
      "最后请补一段「可写回 Markdown 的学习旁注」：像用户写在教材边上的笔记一样，补充更顺手的解释、例子、易错点和复述提示，不要替换原教材。",
    ]
      .filter(Boolean)
      .join("\n");

    setPhase("teaching");
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
    if (!documentId || !objective || !tutorAdvice.trim() || isAppendingTutorNote) return;

    setIsAppendingTutorNote(true);
    try {
      const sourceDocument = await documentApi.getContent(documentId);
      const title = objective.title;
      const note = ["", "---", "", `## 学习旁注：${title}`, "", tutorAdvice.trim(), ""].join("\n");
      const nextContent = `${sourceDocument.content?.trimEnd() || ""}${note}`;

      await documentApi.updateContent(documentId, { content: nextContent });
      await queryClient.invalidateQueries({
        queryKey: ["feynman-source-document", documentId],
      });
      toast.success("学习旁注已追加到 Markdown");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "追加学习旁注失败");
    } finally {
      setIsAppendingTutorNote(false);
    }
  };

  /* 加载骨架 */
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

  /* 目标不存在 */
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
      {/* 工作台头部 */}
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className=" min-w-0 flex flex-wrap items-center gap-2">
          <h1 className="mt-2 truncate text-2xl font-bold">{objective.title}</h1>
          <Badge variant="outline">
            {masteryLabels[objective.mastery_level] ?? objective.mastery_level}
          </Badge>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <PhaseBadge phase={phase} onPhaseChange={setPhase} />
          <Button variant="outline" onClick={() => navigate({ to: "/wiki" })}>
            返回 Wiki
          </Button>
          <Button variant="secondary" onClick={() => navigate({ to: "/learn/feynman" })}>
            返回费曼练习
          </Button>
        </div>
      </div>

      {/* 阅读阶段 */}
      {phase === "reading" ? (
        <SourcePanel
          documentId={documentId}
          objective={objective}
          objectives={objectives}
          isObjectivesLoading={isLoading}
          onOpenObjective={openObjective}
        />
      ) : phase === "answering" ? (
        /* 复述阶段 */
        <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
          <PracticePanel
            answer={answer}
            result={result}
            disabled={!sessionId || submitExercise.isPending}
            isSubmitting={submitExercise.isPending}
            objective={objective}
            onAnswerChange={setAnswer}
            onSubmit={submit}
            onReset={resetAnswer}
            tutorAdvice={tutorAdvice}
            isTutorTeaching={isTutorTeaching}
            onRequestTutorTeaching={requestTutorTeaching}
            canAppendTutorNote={!!documentId}
            isAppendingTutorNote={isAppendingTutorNote}
            onAppendTutorNote={appendTutorNoteToMarkdown}
            onPhaseChange={setPhase}
          />
          <StudyInfoPanel objective={objective} result={result} sessionId={sessionId} />
        </div>
      ) : (
        /* 教学阶段 */
        <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
          <TeachingPanel
            tutorAdvice={tutorAdvice}
            isTutorTeaching={isTutorTeaching}
            canRequestTeaching={!!result}
            canAppendTutorNote={!!documentId}
            isAppendingTutorNote={isAppendingTutorNote}
            onRequestTutorTeaching={requestTutorTeaching}
            onAppendTutorNote={appendTutorNoteToMarkdown}
          />
          <StudyInfoPanel objective={objective} result={result} sessionId={sessionId} />
        </div>
      )}
    </div>
  );
}
