import { useQuery } from "@tanstack/react-query";
import { TableOfContentsIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import { extractMarkdownCatalog, scrollToMarkdownHeading } from "../_shared";

export function SourcePanel({ documentId }: { documentId: string }) {
  const [catalogOpen, setCatalogOpen] = useState(true);
  const {
    data: sourceDocument,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["feynman-source-document", documentId],
    queryFn: () => documentApi.getContent(documentId),
    enabled: !!documentId,
  });
  const sourceMarkdown = sourceDocument?.content?.trim() || "";
  const catalog = useMemo(() => extractMarkdownCatalog(sourceMarkdown), [sourceMarkdown]);

  return (
    <div
      className={cn(
        "grid min-h-0 flex-1 gap-4",
        catalogOpen ? "lg:grid-cols-[240px_minmax(0,1fr)]" : "grid-cols-1",
      )}
    >
      {catalogOpen ? (
        <aside className="hidden min-h-0 overflow-hidden rounded-lg border bg-muted/20 lg:block">
          <div className="flex h-11 items-center gap-2 border-b px-3">
            <TableOfContentsIcon />
            <div className="truncate text-sm font-medium">文章目录</div>
          </div>
          <ScrollArea className="h-[calc(100%-44px)]">
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
                <p className="px-2 py-1.5 text-xs leading-5 text-muted-foreground">
                  当前文章没有可识别标题。
                </p>
              )}
            </nav>
          </ScrollArea>
        </aside>
      ) : null}

      <section className="min-h-0 overflow-hidden rounded-lg border bg-background">
        <div className="flex h-full min-h-0 flex-col">
          <div className="flex h-11 shrink-0 items-center gap-3 border-b px-4">
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={catalogOpen ? "收起目录" : "展开目录"}
              title={catalogOpen ? "收起目录" : "展开目录"}
              onClick={() => setCatalogOpen((open) => !open)}
            >
              <TableOfContentsIcon />
            </Button>
            <div className="min-w-0 truncate text-sm font-medium">
              {sourceDocument?.filename || "文章正文"}
            </div>
          </div>

          <ScrollArea className="min-h-0 flex-1">
            <article className="mx-auto max-w-3xl px-5 py-6 lg:px-8">
              {isLoading ? (
                <div className="flex flex-col gap-3">
                  <Skeleton className="h-6 w-64" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-11/12" />
                  <Skeleton className="h-4 w-4/5" />
                </div>
              ) : sourceMarkdown ? (
                <MessageResponse className="max-w-none text-sm leading-7">
                  {sourceMarkdown}
                </MessageResponse>
              ) : (
                <Empty className="min-h-80 border-0">
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <TableOfContentsIcon />
                    </EmptyMedia>
                    <EmptyTitle>{isError ? "文章读取失败" : "文章没有正文"}</EmptyTitle>
                    <EmptyDescription>
                      {isError
                        ? "暂时无法读取 Markdown 内容，请稍后重试。"
                        : "请先回到 Wiki 为这篇文章补充内容。"}
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </article>
          </ScrollArea>
        </div>
      </section>
    </div>
  );
}
