import { useQuery } from "@tanstack/react-query";
import {
  BookOpenTextIcon,
  ListChecksIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import type { LearningObjective } from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import { extractMarkdownCatalog, scrollToMarkdownHeading } from "../_shared";
import { ObjectiveOutline } from "./objective-outline";

export function SourcePanel({
  documentId,
  objective,
  objectives,
  isObjectivesLoading,
  onOpenObjective,
}: {
  documentId: string;
  objective: LearningObjective;
  objectives: LearningObjective[];
  isObjectivesLoading: boolean;
  onOpenObjective: (id: string) => void;
}) {
  const [catalogOpen, setCatalogOpen] = useState(true);
  const [objectivesOpen, setObjectivesOpen] = useState(true);
  const {
    data: sourceDocument,
    isLoading: isSourceLoading,
    isError: isSourceError,
  } = useQuery({
    queryKey: ["feynman-source-document", documentId],
    queryFn: () => documentApi.getContent(documentId),
    enabled: !!documentId,
  });
  const sourceMarkdown = sourceDocument?.content?.trim() || "";
  const catalog = useMemo(() => extractMarkdownCatalog(sourceMarkdown), [sourceMarkdown]);
  const currentIndex = objectives.findIndex((item) => item.id === objective.id);
  const previousObjective = currentIndex > 0 ? objectives[currentIndex - 1] : null;
  const nextObjective =
    currentIndex >= 0 && currentIndex < objectives.length - 1 ? objectives[currentIndex + 1] : null;
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
              {objectives.length || 1}
            </Badge>
          </div>
          <ScrollArea className="h-[calc(100%-45px)]">
            <ObjectiveOutline
              objective={objective}
              objectives={objectives}
              isLoading={isObjectivesLoading}
              previousObjective={previousObjective}
              nextObjective={nextObjective}
              onOpenObjective={onOpenObjective}
            />
          </ScrollArea>
        </aside>
      ) : null}
    </div>
  );
}
