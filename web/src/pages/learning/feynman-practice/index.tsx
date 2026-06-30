import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import {
  ArrowLeftIcon,
  BookOpenTextIcon,
  CheckCircle2Icon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
  CircleAlertIcon,
  CircleDashedIcon,
  GraduationCapIcon,
  ListChecksIcon,
  type LucideIcon,
  PlayCircleIcon,
  RotateCcwIcon,
  RouteIcon,
  TargetIcon,
} from "lucide-react";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import {
  sessionChatStream,
  useCreateSession,
  useGenerateGuide,
  useGuideCache,
  useGoalDetail,
  useSubmitExercise,
  guideKeys,
  type ExerciseResult,
  type GuidePracticePoint,
  type GuideResult,
  type LearningObjective,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import { FeynmanAnswerEditor } from "./_components/feynman-answer-editor";

type WorkbenchPhase = "reading" | "answering";

type MarkdownCatalogItem = {
  line: number;
  level: number;
  text: string;
};

type GuideContent = {
  summary: string;
  focusItems: string[];
  practicePoints: GuidePracticePoint[];
  readingSteps: string[];
  pitfalls: string[];
  selfCheckQuestions: string[];
  evidenceItems: string[];
};

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
  const queryClient = useQueryClient();

  const [sessionId, setSessionId] = useState("");
  const [phase, setPhase] = useState<WorkbenchPhase>("reading");
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState<ExerciseResult | null>(null);
  const [tutorAdvice, setTutorAdvice] = useState("");
  const [isTutorTeaching, setIsTutorTeaching] = useState(false);
  const [selectedPracticePoint, setSelectedPracticePoint] = useState<GuidePracticePoint | null>(
    null,
  );
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
    setTutorAdvice("");
    setIsTutorTeaching(false);
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
    void queryClient.invalidateQueries({ queryKey: ["learning-profiles", goalId] });
  };

  const resetAnswer = () => {
    setAnswer("");
    setResult(null);
    setTutorAdvice("");
    setIsTutorTeaching(false);
  };

  const requestTutorTeaching = async () => {
    if (!sessionId || !objective || !result || isTutorTeaching) return;

    const message = [
      `我刚才复述「${selectedPracticePoint?.title || objective.title}」没有通过。`,
      selectedPracticePoint?.goal ? `本轮目标是：${selectedPracticePoint.goal}` : "",
      `Examiner 的反馈是：${result.feedback}`,
      "请你不要只评价我，直接教我：先用通俗的话讲清楚这个知识点，再指出我漏掉的关键点，最后给我一个很小的复述练习。",
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
          />
          <StudyInfoPanel objective={objective} result={result} sessionId={sessionId} />
        </div>
      )}
    </div>
  );
}

function PhaseBadge({
  phase,
  onPhaseChange,
}: {
  phase: WorkbenchPhase;
  onPhaseChange: (phase: WorkbenchPhase) => void;
}) {
  return (
    <div className="flex items-center gap-1 rounded-md border bg-background p-1 text-xs">
      <button
        type="button"
        className={`rounded px-2 py-1 ${
          phase === "reading" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
        onClick={() => onPhaseChange("reading")}
      >
        1 阅读
      </button>
      <button
        type="button"
        className={`rounded px-2 py-1 ${
          phase === "answering" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
        onClick={() => onPhaseChange("answering")}
      >
        2 复述
      </button>
    </div>
  );
}

function SourcePanel({
  objective,
  stageObjectives,
  selectedPracticePoint,
  onPracticePointSelect,
}: {
  objective: LearningObjective;
  stageObjectives: LearningObjective[];
  selectedPracticePoint: GuidePracticePoint | null;
  onPracticePointSelect: (point: GuidePracticePoint | null) => void;
}) {
  const documentId = objective.source_document_id;
  const [catalogOpen, setCatalogOpen] = useState(true);
  const [guideOpen, setGuideOpen] = useState(true);
  const {
    data: sourceDocument,
    isLoading: isSourceLoading,
    isError: isSourceError,
  } = useQuery({
    queryKey: ["feynman-source-document", documentId],
    queryFn: () => documentApi.getContent(documentId as string),
    enabled: !!documentId,
  });
  const sourceMarkdown = sourceDocument?.content?.trim() || "";
  const contentHash = useSHA256(sourceMarkdown);
  const catalog = useMemo(() => extractMarkdownCatalog(sourceMarkdown), [sourceMarkdown]);
  const generateGuide = useGenerateGuide();
  const queryClient = useQueryClient();
  const {
    data: cachedGuideData,
    isLoading: isGuideCacheLoading,
    isError: isGuideCacheError,
  } = useGuideCache(objective.id, contentHash);
  const [guideObjectiveId, setGuideObjectiveId] = useState("");
  const {
    data: guideData,
    isError: isGuideError,
    isPending: isGuidePending,
    mutate: generateGuideMutate,
    reset: resetGuide,
  } = generateGuide;
  const generatedGuideData = guideObjectiveId === objective.id ? guideData : null;
  const currentGuideData = cachedGuideData ?? generatedGuideData ?? null;

  useEffect(() => {
    setGuideObjectiveId("");
    resetGuide();
  }, [objective.id, resetGuide]);

  useEffect(() => {
    const practicePoints = toArray(currentGuideData?.practice_points);
    if (practicePoints.length === 0 || selectedPracticePoint) return;
    onPracticePointSelect(practicePoints[0]);
  }, [currentGuideData?.practice_points, onPracticePointSelect, selectedPracticePoint]);

  const requestGuide = () => {
    if (!objective.id || !sourceMarkdown || isGuidePending) return;
    setGuideObjectiveId(objective.id);
    generateGuideMutate(
      { objective_id: objective.id, markdown: sourceMarkdown },
      {
        onSuccess: (data) => {
          if (!data.content_hash) return;
          queryClient.setQueryData(guideKeys.detail(objective.id, data.content_hash), data);
        },
      },
    );
  };
  const readingGridClassName = cn(
    "grid min-h-0 flex-1 gap-4",
    catalogOpen && guideOpen
      ? "lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1fr)_340px]"
      : catalogOpen
        ? "lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1fr)]"
        : guideOpen
          ? "lg:grid-cols-[minmax(0,1fr)] xl:grid-cols-[minmax(0,1fr)_340px]"
          : "lg:grid-cols-[minmax(0,1fr)]",
  );

  return (
    <div className={readingGridClassName}>
      {catalogOpen ? (
        <aside className="hidden min-h-0 overflow-hidden rounded-2xl border bg-muted/20 lg:block">
          <div className="flex h-[45px] items-center gap-2 border-b px-3">
            <BookOpenTextIcon className="size-4 text-muted-foreground" />
            <div className="truncate text-sm font-medium">目录</div>
          </div>
          <ScrollArea className="h-[calc(100%-45px)]">
            <nav className="flex flex-col gap-1 p-3 text-sm">
              {catalog.length > 0 ? (
                catalog.map((item) => (
                  <button
                    key={`${item.line}-${item.text}`}
                    type="button"
                    className="block w-full truncate rounded-md px-2 py-1.5 text-left text-muted-foreground transition-colors hover:bg-background hover:text-foreground"
                    style={{ paddingLeft: `${8 + (item.level - 1) * 12}px` }}
                    onClick={() => scrollToMarkdownHeading(item.text)}
                  >
                    {item.text}
                  </button>
                ))
              ) : (
                <div className="px-2 py-1.5 text-xs leading-5 text-muted-foreground">
                  当前文档没有可识别标题。
                </div>
              )}
            </nav>
          </ScrollArea>
        </aside>
      ) : null}

      <section className="min-h-0 overflow-hidden rounded-2xl border bg-background">
        <div className="flex h-full min-h-0 flex-col">
          <div className="flex shrink-0 items-center justify-between gap-3 border-b px-5 py-3">
            <div className="flex min-w-0 items-center gap-3">
              <Button
                variant="ghost"
                size="icon"
                className="hidden size-8 shrink-0 lg:inline-flex"
                onClick={() => setCatalogOpen((open) => !open)}
              >
                {catalogOpen ? (
                  <PanelLeftCloseIcon className="size-4" />
                ) : (
                  <PanelLeftOpenIcon className="size-4" />
                )}
              </Button>
              <div className="min-w-0">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <BookOpenTextIcon className="size-3.5" />
                  Markdown 阅读
                </div>
                <div className="mt-1 truncate text-base font-semibold">{objective.title}</div>
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                className="size-8"
                onClick={() => setGuideOpen((open) => !open)}
              >
                {guideOpen ? (
                  <PanelRightCloseIcon className="size-4" />
                ) : (
                  <PanelRightOpenIcon className="size-4" />
                )}
              </Button>
            </div>
          </div>

          <ScrollArea className="min-h-0 flex-1">
            <article className="mx-auto max-w-3xl px-5 py-6 lg:px-8">
              {isSourceLoading ? (
                <div className="space-y-3">
                  <Skeleton className="h-5 w-64" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-11/12" />
                  <Skeleton className="h-4 w-4/5" />
                </div>
              ) : sourceMarkdown ? (
                <MessageResponse className="max-w-none text-sm leading-7">
                  {sourceMarkdown}
                </MessageResponse>
              ) : (
                <div className="text-sm leading-7 text-muted-foreground">
                  {isSourceError
                    ? "原始 Markdown 文档读取失败，请稍后重试。"
                    : "这个小目标还没有关联原始 Markdown 文档。"}
                </div>
              )}
            </article>
          </ScrollArea>
        </div>
      </section>

      {guideOpen ? (
        <aside className="min-h-0 overflow-hidden rounded-2xl border bg-muted/10">
          <div className="flex h-[45px] items-center gap-2 border-b px-3">
            <GraduationCapIcon className="size-4 text-muted-foreground" />
            <div className="truncate text-sm font-medium">导学 Agent</div>
          </div>
          <ScrollArea className="h-[calc(100%-45px)]">
            <TeachingGuideAgent
              objective={objective}
              stageObjectives={stageObjectives}
              guideResult={currentGuideData ?? null}
              guideStatus={
                isGuidePending
                  ? "loading"
                  : isGuideError && guideObjectiveId === objective.id
                    ? "error"
                    : isGuideCacheLoading
                      ? "cache-loading"
                      : isGuideCacheError
                        ? "cache-error"
                        : currentGuideData
                          ? "agent"
                          : "idle"
              }
              selectedPracticePoint={selectedPracticePoint}
              onPracticePointSelect={onPracticePointSelect}
              canGenerate={!!sourceMarkdown && !isSourceLoading}
              onGenerate={requestGuide}
            />
          </ScrollArea>
        </aside>
      ) : null}
    </div>
  );
}

function TeachingGuideAgent({
  objective,
  stageObjectives,
  guideResult,
  guideStatus,
  selectedPracticePoint,
  onPracticePointSelect,
  canGenerate,
  onGenerate,
}: {
  objective: LearningObjective;
  stageObjectives: LearningObjective[];
  guideResult: GuideResult | null;
  guideStatus: "idle" | "loading" | "cache-loading" | "cache-error" | "agent" | "error";
  selectedPracticePoint: GuidePracticePoint | null;
  onPracticePointSelect: (point: GuidePracticePoint) => void;
  canGenerate: boolean;
  onGenerate: () => void;
}) {
  const currentIndex = stageObjectives.findIndex((item) => item.id === objective.id);
  const previousObjective = currentIndex > 0 ? stageObjectives[currentIndex - 1] : null;
  const nextObjective =
    currentIndex >= 0 && currentIndex < stageObjectives.length - 1
      ? stageObjectives[currentIndex + 1]
      : null;
  const guideContent = guideResult ? guideResultToContent(guideResult) : null;
  const isAgentGuide = guideStatus === "agent";

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between gap-2 rounded-md bg-muted/50 px-3 py-2">
        <div className="flex min-w-0 items-center gap-2">
          <GraduationCapIcon className="size-4 text-primary" />
          <span className="truncate text-sm font-medium">
            {guideStatus === "loading"
              ? "Generating..."
              : guideStatus === "cache-loading"
                ? "读取缓存..."
                : isAgentGuide
                  ? guideResult?.cached
                    ? "Saved Guide"
                    : "Real Agent"
                  : "等待生成"}
          </span>
        </div>
        <Badge variant={isAgentGuide ? "default" : "secondary"}>
          {isAgentGuide ? (guideResult?.cached ? "已保存" : "真实") : "手动"}
        </Badge>
      </div>

      {!guideContent ? (
        <div className="flex min-h-72 flex-col items-center justify-center gap-4 rounded-md border bg-background p-4 text-center">
          <GraduationCapIcon className="size-8 text-muted-foreground" />
          <div className="flex flex-col gap-2">
            <div className="text-sm font-medium">
              {guideStatus === "error" ? "导学 Agent 生成失败" : "点击后生成导学内容"}
            </div>
            <p className="max-w-64 text-sm leading-6 text-muted-foreground">
              {guideStatus === "error"
                ? "请稍后重试，或先阅读原文继续复述。"
                : "Agent 会阅读当前 Markdown，并保存掌握目标、复述小点和自检问题。"}
            </p>
          </div>
          <Button onClick={onGenerate} disabled={!canGenerate || guideStatus === "loading"}>
            <GraduationCapIcon className="size-4" />
            {guideStatus === "loading" ? "生成中..." : "生成导学"}
          </Button>
        </div>
      ) : null}

      {guideContent ? (
        <>
          <section className="flex flex-col gap-2">
            <div className="flex items-center gap-2 text-sm font-medium">
              <GraduationCapIcon className="size-4 text-primary" />
              本节要掌握
            </div>
            <p className="text-sm leading-6 text-muted-foreground">
              {guideStatus === "loading"
                ? "真实导学 agent 正在阅读资料并生成本节目标。"
                : guideContent.summary ||
                  `先把「${objective.title}」讲清楚：它解决什么问题、核心概念是什么、和本阶段前后知识点怎么接上。`}
            </p>
          </section>

          <div className="flex justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={onGenerate}
              disabled={!canGenerate || guideStatus === "loading"}
            >
              <GraduationCapIcon className="size-4" />
              {guideStatus === "loading" ? "生成中..." : "重新生成"}
            </Button>
          </div>

          <GuideSection icon={TargetIcon} title={isAgentGuide ? "Agent 掌握目标" : "掌握目标"}>
            <ul className="flex flex-col gap-2 text-sm leading-6 text-muted-foreground">
              {guideContent.focusItems.map((item) => (
                <li key={item} className="flex gap-2">
                  <span className="mt-2 size-1.5 shrink-0 rounded-full bg-primary" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </GuideSection>

          {guideContent.practicePoints.length > 0 ? (
            <GuideSection icon={PlayCircleIcon} title="本轮复述小点">
              <div className="flex flex-col gap-2">
                {guideContent.practicePoints.map((point, index) => {
                  const selected = selectedPracticePoint?.title === point.title;
                  return (
                    <button
                      key={`${point.title}-${index}`}
                      type="button"
                      className={cn(
                        "w-full rounded-md border bg-background px-3 py-2 text-left transition-colors hover:border-primary/60 hover:bg-primary/5",
                        selected && "border-primary bg-primary/10",
                      )}
                      onClick={() => onPracticePointSelect(point)}
                    >
                      <div className="flex items-start gap-2">
                        <span
                          className={cn(
                            "mt-1 flex size-5 shrink-0 items-center justify-center rounded-full border text-xs font-medium",
                            selected
                              ? "border-primary bg-primary text-primary-foreground"
                              : "bg-muted text-muted-foreground",
                          )}
                        >
                          {index + 1}
                        </span>
                        <span className="min-w-0">
                          <span className="block text-sm font-medium leading-5">{point.title}</span>
                          {point.goal ? (
                            <span className="mt-1 block text-xs leading-5 text-muted-foreground">
                              {point.goal}
                            </span>
                          ) : null}
                        </span>
                      </div>
                    </button>
                  );
                })}
              </div>
            </GuideSection>
          ) : null}

          <GuideSection icon={ListChecksIcon} title="阅读顺序">
            <ol className="flex flex-col gap-2 text-sm leading-6 text-muted-foreground">
              {guideContent.readingSteps.map((item, index) => (
                <li key={item} className="flex gap-2">
                  <span className="flex size-5 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-medium text-foreground">
                    {index + 1}
                  </span>
                  <span>{item}</span>
                </li>
              ))}
            </ol>
          </GuideSection>

          {guideContent.pitfalls.length > 0 ? (
            <GuideSection icon={CircleAlertIcon} title="易错点">
              <ul className="flex flex-col gap-2 text-sm leading-6 text-muted-foreground">
                {guideContent.pitfalls.map((item) => (
                  <li key={item} className="flex gap-2">
                    <span className="mt-2 size-1.5 shrink-0 rounded-full bg-primary" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </GuideSection>
          ) : null}

          <GuideSection icon={RouteIcon} title="阶段位置">
            <div className="flex flex-col gap-2 text-sm">
              <GuideContextRow
                label="上一小节"
                value={previousObjective?.title || "这是本阶段开头"}
              />
              <GuideContextRow label="当前小节" value={objective.title} active />
              <GuideContextRow
                label="下一小节"
                value={nextObjective?.title || "读完后进入复述验证"}
              />
            </div>
          </GuideSection>

          <GuideSection icon={BookOpenTextIcon} title="资料依据">
            <div className="flex flex-col gap-2 text-sm leading-6 text-muted-foreground">
              {guideContent.evidenceItems.map((item) => (
                <blockquote key={item} className="border-l-2 pl-3">
                  {item}
                </blockquote>
              ))}
            </div>
          </GuideSection>

          <div className="rounded-md border bg-background p-3">
            <div className="mb-2 text-sm font-medium">复述前自检</div>
            {guideContent.selfCheckQuestions.length > 0 ? (
              <ul className="flex flex-col gap-2 text-sm leading-6 text-muted-foreground">
                {guideContent.selfCheckQuestions.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            ) : (
              <p className="text-sm leading-6 text-muted-foreground">
                如果你能不用原文解释「是什么、为什么、怎么用、哪里容易错」，就可以进入费曼复述。
              </p>
            )}
          </div>
        </>
      ) : null}
    </div>
  );
}

function GuideSection({
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

function GuideContextRow({
  label,
  value,
  active,
}: {
  label: string;
  value: string;
  active?: boolean;
}) {
  return (
    <div className={`rounded-md border p-2 ${active ? "border-primary bg-primary/5" : ""}`}>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 line-clamp-2 font-medium leading-5">{value}</div>
    </div>
  );
}

function PracticePanel({
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
    <Card className="min-h-0 overflow-hidden rounded-2xl py-0">
      <CardHeader className="shrink-0 border-b p-4!">
        <CardTitle className="text-base">本次学习信息</CardTitle>
      </CardHeader>
      <CardContent className="min-h-0 flex-1 p-0">
        <ScrollArea className="h-full">
          <div className="space-y-4 p-4">
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
            {result ? (
              <div className="space-y-3 rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                <InfoBlock label="判定依据" value={result.evidence} fallback="本次未返回判定依据" />
                <InfoBlock
                  label="薄弱点"
                  value={
                    result.weak_points && result.weak_points.length > 0
                      ? result.weak_points.join("、")
                      : "暂无明显薄弱点"
                  }
                />
                <InfoBlock
                  label="下一步"
                  value={result.next_recommendation}
                  fallback="继续推进下一个小目标"
                />
                <div className="flex items-center justify-between gap-3 border-t pt-3">
                  <span>复习标记</span>
                  <Badge variant={result.review_required ? "secondary" : "outline"}>
                    {result.review_required ? "需要复习" : "暂不需要"}
                  </Badge>
                </div>
              </div>
            ) : (
              <div className="rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                提交解释后，Learning Examiner 会同步写入练习记录、学习画像和本次日志。
              </div>
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

function InfoBlock({
  label,
  value,
  fallback = "-",
}: {
  label: string;
  value?: string;
  fallback?: string;
}) {
  return (
    <div className="space-y-1">
      <div className="text-xs font-medium text-foreground">{label}</div>
      <div className="break-words">{value || fallback}</div>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="shrink-0 text-muted-foreground">{label}</span>
      <span className="min-w-0 max-w-44 truncate text-right font-medium">{value}</span>
    </div>
  );
}

function extractMarkdownCatalog(markdown: string): MarkdownCatalogItem[] {
  return markdown
    .split("\n")
    .map((line, index) => {
      const match = /^(#{1,4})\s+(.+)$/.exec(line.trim());
      if (!match) return null;
      return {
        line: index,
        level: match[1].length,
        text: match[2].replace(/[#*_`]/g, "").trim(),
      };
    })
    .filter((item): item is MarkdownCatalogItem => !!item && item.text.length > 0);
}

function useSHA256(text: string) {
  const [hash, setHash] = useState("");

  useEffect(() => {
    let cancelled = false;
    const value = text.trim();
    if (!value) {
      setHash("");
      return;
    }

    crypto.subtle
      .digest("SHA-256", new TextEncoder().encode(value))
      .then((buffer) => {
        if (cancelled) return;
        const bytes = Array.from(new Uint8Array(buffer));
        setHash(bytes.map((byte) => byte.toString(16).padStart(2, "0")).join(""));
      })
      .catch(() => {
        if (!cancelled) setHash("");
      });

    return () => {
      cancelled = true;
    };
  }, [text]);

  return hash;
}

function guideResultToContent(result: GuideResult): GuideContent {
  return {
    summary: result.summary,
    focusItems: toArray(result.mastery_goals),
    practicePoints: toArray(result.practice_points),
    readingSteps: toArray(result.reading_steps),
    pitfalls: toArray(result.pitfalls),
    selfCheckQuestions: toArray(result.self_check_questions),
    evidenceItems: toArray(result.evidence),
  };
}

function toArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

function scrollToMarkdownHeading(text: string) {
  const headings = Array.from(
    document.querySelectorAll<HTMLElement>("article h1, article h2, article h3, article h4"),
  );
  const heading = headings.find((item) => item.textContent?.trim() === text);
  heading?.scrollIntoView({ block: "start", behavior: "smooth" });
}

function buildPrompt(objective: LearningObjective, practicePoint: GuidePracticePoint | null) {
  if (!practicePoint) {
    return `请用自己的话解释：${objective.title}`;
  }

  return [
    `请用自己的话解释：${objective.title}`,
    `本轮只判断这个复述小点：${practicePoint.title}`,
    practicePoint.goal ? `本轮目标：${practicePoint.goal}` : "",
    "请不要按整篇资料要求判定，只看这个小点是否讲清楚。",
  ]
    .filter(Boolean)
    .join("\n");
}
