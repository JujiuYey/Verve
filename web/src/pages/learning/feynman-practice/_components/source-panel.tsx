import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  BookOpenTextIcon,
  CircleAlertIcon,
  GraduationCapIcon,
  ListChecksIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
  PlayCircleIcon,
  RouteIcon,
  TargetIcon,
  type LucideIcon,
} from "lucide-react";
import { type ReactNode, useEffect, useMemo, useState } from "react";

import {
  guideKeys,
  useGenerateGuide,
  useGuideCache,
  type GuidePracticePoint,
  type GuideResult,
  type LearningObjective,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import {
  extractMarkdownCatalog,
  guideResultToContent,
  scrollToMarkdownHeading,
  toArray,
} from "../_shared";
import { useSHA256 } from "../_hooks/use-sha-256";

export function SourcePanel({
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

