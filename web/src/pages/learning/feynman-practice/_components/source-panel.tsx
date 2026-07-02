import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
  BookOpenTextIcon,
  CircleAlertIcon,
  ListChecksIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
  RouteIcon,
  TargetIcon,
  type LucideIcon,
} from "lucide-react";
import { type ReactNode, useMemo, useState } from "react";

import { useObjectives, type LearningObjective } from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import { extractMarkdownCatalog, masteryLabels, scrollToMarkdownHeading } from "../_shared";

export function SourcePanel({ objective }: { objective: LearningObjective }) {
  const navigate = useNavigate();
  const documentId = objective.source_document_id;
  const [catalogOpen, setCatalogOpen] = useState(true);
  const [objectivesOpen, setObjectivesOpen] = useState(true);
  const {
    data: sourceDocument,
    isLoading: isSourceLoading,
    isError: isSourceError,
  } = useQuery({
    queryKey: ["feynman-source-document", documentId],
    queryFn: () => documentApi.getContent(documentId as string),
    enabled: !!documentId,
  });
  const { data: documentObjectives = [], isLoading: isObjectivesLoading } = useObjectives({
    document_id: documentId,
  });
  const sourceMarkdown = sourceDocument?.content?.trim() || "";
  const catalog = useMemo(() => extractMarkdownCatalog(sourceMarkdown), [sourceMarkdown]);
  const currentIndex = documentObjectives.findIndex((item) => item.id === objective.id);
  const previousObjective = currentIndex > 0 ? documentObjectives[currentIndex - 1] : null;
  const nextObjective =
    currentIndex >= 0 && currentIndex < documentObjectives.length - 1
      ? documentObjectives[currentIndex + 1]
      : null;
  const readingGridClassName = cn(
    "grid min-h-0 flex-1 gap-4",
    catalogOpen && objectivesOpen
      ? "lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1fr)_340px]"
      : catalogOpen
        ? "lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1fr)]"
        : objectivesOpen
          ? "lg:grid-cols-[minmax(0,1fr)] xl:grid-cols-[minmax(0,1fr)_340px]"
          : "lg:grid-cols-[minmax(0,1fr)]",
  );

  const openObjective = (id: string) => {
    if (id === objective.id) return;
    navigate({
      to: "/learn/feynman-practice/$objectiveId",
      params: { objectiveId: id },
    });
  };

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
            <Button
              variant="ghost"
              size="icon"
              className="size-8 shrink-0"
              onClick={() => setObjectivesOpen((open) => !open)}
            >
              {objectivesOpen ? (
                <PanelRightCloseIcon className="size-4" />
              ) : (
                <PanelRightOpenIcon className="size-4" />
              )}
            </Button>
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
                    : "这个小节还没有关联原始 Markdown 文档。"}
                </div>
              )}
            </article>
          </ScrollArea>
        </div>
      </section>

      {objectivesOpen ? (
        <aside className="min-h-0 overflow-hidden rounded-2xl border bg-muted/10">
          <div className="flex h-[45px] items-center gap-2 border-b px-3">
            <ListChecksIcon className="size-4 text-muted-foreground" />
            <div className="truncate text-sm font-medium">学习小节</div>
            <Badge variant="outline" className="ml-auto">
              {documentObjectives.length || 1}
            </Badge>
          </div>
          <ScrollArea className="h-[calc(100%-45px)]">
            <ObjectiveOutline
              objective={objective}
              objectives={documentObjectives}
              isLoading={isObjectivesLoading}
              previousObjective={previousObjective}
              nextObjective={nextObjective}
              onOpenObjective={openObjective}
            />
          </ScrollArea>
        </aside>
      ) : null}
    </div>
  );
}

function ObjectiveOutline({
  objective,
  objectives,
  isLoading,
  previousObjective,
  nextObjective,
  onOpenObjective,
}: {
  objective: LearningObjective;
  objectives: LearningObjective[];
  isLoading: boolean;
  previousObjective: LearningObjective | null;
  nextObjective: LearningObjective | null;
  onOpenObjective: (id: string) => void;
}) {
  const items = objectives.length > 0 ? objectives : [objective];

  return (
    <div className="flex flex-col gap-4 p-4">
      <Section icon={TargetIcon} title="当前要掌握">
        <div className="rounded-md border border-primary/30 bg-primary/5 p-3">
          <div className="text-sm font-medium leading-5">{objective.title}</div>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {objective.detail || "用自己的话讲清这个小节的定义、用途、边界和易错点。"}
          </p>
          <Badge variant="outline" className="mt-3">
            {masteryLabels[objective.mastery_level] ?? objective.mastery_level}
          </Badge>
        </div>
      </Section>

      <Section icon={ListChecksIcon} title="文档小节">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {items.map((item, index) => {
              const active = item.id === objective.id;
              return (
                <button
                  key={item.id}
                  type="button"
                  className={cn(
                    "w-full rounded-md border bg-background px-3 py-2 text-left transition-colors hover:border-primary/60 hover:bg-primary/5",
                    active && "border-primary bg-primary/10",
                  )}
                  onClick={() => onOpenObjective(item.id)}
                >
                  <div className="flex items-start gap-2">
                    <span
                      className={cn(
                        "mt-1 flex size-5 shrink-0 items-center justify-center rounded-full border text-xs font-medium",
                        active
                          ? "border-primary bg-primary text-primary-foreground"
                          : "bg-muted text-muted-foreground",
                      )}
                    >
                      {index + 1}
                    </span>
                    <span className="min-w-0">
                      <span className="block text-sm font-medium leading-5">{item.title}</span>
                      {item.detail ? (
                        <span className="mt-1 line-clamp-3 block text-xs leading-5 text-muted-foreground">
                          {item.detail}
                        </span>
                      ) : null}
                    </span>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </Section>

      <Section icon={RouteIcon} title="阶段位置">
        <div className="flex flex-col gap-2 text-sm">
          <ContextRow label="上一小节" value={previousObjective?.title || "这是当前文档开头"} />
          <ContextRow label="当前小节" value={objective.title} active />
          <ContextRow label="下一小节" value={nextObjective?.title || "读完后进入复述验证"} />
        </div>
      </Section>

      <div className="rounded-md border bg-background p-3">
        <div className="mb-2 flex items-center gap-2 text-sm font-medium">
          <CircleAlertIcon className="size-4 text-primary" />
          复述前自检
        </div>
        <p className="text-sm leading-6 text-muted-foreground">
          先只围绕当前小节复述。能讲出是什么、为什么、怎么用、哪里容易错，再进入复述验证。
        </p>
      </div>
    </div>
  );
}

function Section({
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

function ContextRow({ label, value, active }: { label: string; value: string; active?: boolean }) {
  return (
    <div className={`rounded-md border p-2 ${active ? "border-primary bg-primary/5" : ""}`}>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 line-clamp-2 font-medium leading-5">{value}</div>
    </div>
  );
}
