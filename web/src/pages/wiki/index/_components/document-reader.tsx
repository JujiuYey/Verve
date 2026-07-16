import { useQuery } from "@tanstack/react-query";
import {
  AlertCircleIcon,
  BookOpenIcon,
  CheckCircle2Icon,
  ChevronDownIcon,
  Clock3Icon,
  DatabaseIcon,
  DownloadIcon,
  FilePenIcon,
  FileTextIcon,
  GraduationCapIcon,
  Loader2Icon,
  MessageSquareTextIcon,
  RefreshCwIcon,
  TableOfContentsIcon,
  Trash2Icon,
} from "lucide-react";
import { useMemo, useRef, useState } from "react";
import { toast } from "sonner";

import type { LearningAgentType } from "@/api/learning";
import { type IndexJobProgress, type IndexJobStatus, ragWikiApi } from "@/api/rag/wiki";
import type { Document } from "@/api/wiki/document";
import { documentApi } from "@/api/wiki/document";
import { MessageResponse } from "@/components/ai-elements/message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

type CatalogItem = {
  level: number;
  line: number;
  text: string;
};

interface DocumentReaderProps {
  document?: Document;
  indexJob?: IndexJobProgress;
  onDelete: (document: Document) => void;
  onIndexStatusRefresh: () => void;
  onOpenPractice: (document: Document, agentType: LearningAgentType) => void;
}

const statusMeta: Record<
  IndexJobStatus | "not_started",
  {
    icon: React.ComponentType<{ className?: string }>;
    label: string;
    variant: "secondary" | "outline" | "destructive";
  }
> = {
  not_started: { icon: Clock3Icon, label: "未解析", variant: "outline" },
  pending: { icon: Clock3Icon, label: "等待解析", variant: "secondary" },
  running: { icon: Loader2Icon, label: "解析中", variant: "secondary" },
  completed: { icon: CheckCircle2Icon, label: "已解析", variant: "secondary" },
  failed: { icon: AlertCircleIcon, label: "解析失败", variant: "destructive" },
  superseded: { icon: Clock3Icon, label: "版本已过期", variant: "outline" },
};

function extractCatalog(markdown: string): CatalogItem[] {
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
    .filter((item): item is CatalogItem => !!item && item.text.length > 0);
}

function formatFileSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function formatUpdatedAt(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function DocumentReader({
  document,
  indexJob,
  onDelete,
  onIndexStatusRefresh,
  onOpenPractice,
}: DocumentReaderProps) {
  const [catalogOpen, setCatalogOpen] = useState(true);
  const [downloading, setDownloading] = useState(false);
  const [indexing, setIndexing] = useState(false);
  const articleRef = useRef<HTMLElement>(null);
  const documentId = document?.id ?? "";

  const {
    data: contentResponse,
    isError,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ["wiki-document-content", documentId, document?.current_version],
    queryFn: () => documentApi.getContent(documentId),
    enabled: !!documentId,
  });

  const markdown = contentResponse?.content?.trim() ?? "";
  const catalog = useMemo(() => extractCatalog(markdown), [markdown]);
  const currentStatus = indexing ? "running" : (indexJob?.status ?? "not_started");
  const currentStatusMeta = statusMeta[currentStatus];
  const StatusIcon = currentStatusMeta.icon;
  const indexDisabled = indexing || currentStatus === "pending" || currentStatus === "running";
  const indexLabel =
    currentStatus === "not_started"
      ? "解析文档"
      : currentStatus === "pending"
        ? "等待解析"
        : currentStatus === "running"
          ? "解析中"
          : "重新解析";

  const handleCatalogNavigate = (text: string) => {
    const headings = Array.from(
      articleRef.current?.querySelectorAll<HTMLElement>("h1, h2, h3, h4") ?? [],
    );
    headings
      .find((heading) => heading.textContent?.trim() === text)
      ?.scrollIntoView({
        block: "start",
        behavior: "smooth",
      });
  };

  const handleDownload = async () => {
    if (!document) return;
    setDownloading(true);
    try {
      const result = await documentApi.download(document.id);
      window.open(result.download_url, "_blank");
    } catch (error) {
      console.error("下载文档失败:", error);
      toast.error("下载失败");
    } finally {
      setDownloading(false);
    }
  };

  const handleIndex = async () => {
    if (!document || indexDisabled) return;
    setIndexing(true);
    try {
      await ragWikiApi.indexDocument(document.id);
      toast.success("文档解析已提交");
    } catch (error) {
      console.error("解析文档失败:", error);
      toast.error("文档解析失败");
    } finally {
      setIndexing(false);
      onIndexStatusRefresh();
    }
  };

  if (!document) {
    return (
      <Empty className="h-full min-h-80 rounded-none border-0">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <BookOpenIcon />
          </EmptyMedia>
          <EmptyTitle>选择一篇文档开始阅读</EmptyTitle>
          <EmptyDescription>从左侧文件树中选择 Markdown 文档，正文会在这里打开。</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <section className="flex h-full min-h-0 flex-col bg-background">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b px-4 py-3 lg:px-5">
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-2">
            <h1 className="truncate text-base font-semibold" title={document.filename}>
              {document.filename}
            </h1>
            <Badge variant={currentStatusMeta.variant} title={indexJob?.error_message}>
              <StatusIcon className={cn(currentStatus === "running" && "animate-spin")} />
              {currentStatusMeta.label}
              {currentStatus === "completed" && indexJob?.chunk_count
                ? ` · ${indexJob.chunk_count} 片段`
                : null}
            </Badge>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            {formatFileSize(document.file_size)} · 版本 {document.current_version} · 更新于{" "}
            {formatUpdatedAt(document.updated_at)}
          </p>
        </div>

        <TooltipProvider>
          <div className="flex shrink-0 items-center gap-1.5">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="hidden xl:inline-flex"
              aria-label={catalogOpen ? "收起文章目录" : "展开文章目录"}
              onClick={() => setCatalogOpen((open) => !open)}
            >
              <TableOfContentsIcon />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button type="button" variant="outline" size="sm">
                  <GraduationCapIcon data-icon="inline-start" />
                  费曼练习
                  <ChevronDownIcon data-icon="inline-end" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuGroup>
                  <DropdownMenuItem onSelect={() => onOpenPractice(document, "listener")}>
                    <MessageSquareTextIcon />
                    开始讲解
                  </DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => onOpenPractice(document, "teacher")}>
                    <GraduationCapIcon />
                    教学补充
                  </DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => onOpenPractice(document, "curator")}>
                    <FilePenIcon />
                    修订文档
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={indexDisabled}
              onClick={handleIndex}
            >
              {indexing ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <DatabaseIcon data-icon="inline-start" />
              )}
              {indexLabel}
            </Button>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label="下载文档"
                  disabled={downloading}
                  onClick={handleDownload}
                >
                  {downloading ? <Spinner /> : <DownloadIcon />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>下载文档</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label="删除文档"
                  onClick={() => onDelete(document)}
                >
                  <Trash2Icon />
                </Button>
              </TooltipTrigger>
              <TooltipContent>删除文档</TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>
      </header>

      <div
        className={cn(
          "grid min-h-0 flex-1",
          catalogOpen ? "xl:grid-cols-[220px_minmax(0,1fr)]" : "grid-cols-1",
        )}
      >
        {catalogOpen ? (
          <aside className="hidden min-h-0 overflow-hidden border-r bg-muted/20 xl:block">
            <div className="flex h-11 items-center gap-2 border-b px-3 text-sm font-medium">
              <TableOfContentsIcon className="size-4 text-muted-foreground" />
              文章目录
            </div>
            <ScrollArea className="h-[calc(100%-44px)]">
              <nav className="flex flex-col gap-1 p-3 text-sm" aria-label="文章目录">
                {catalog.length > 0 ? (
                  catalog.map((item) => (
                    <button
                      key={`${item.line}-${item.text}`}
                      type="button"
                      className="block w-full truncate rounded-md px-2 py-1.5 text-left text-muted-foreground outline-none transition-colors hover:bg-background hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50"
                      style={{ paddingLeft: `${8 + (item.level - 1) * 12}px` }}
                      onClick={() => handleCatalogNavigate(item.text)}
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

        <ScrollArea className="min-h-0 bg-background">
          <article ref={articleRef} className="mx-auto max-w-3xl px-5 py-7 lg:px-8 lg:py-9">
            {isLoading ? (
              <div className="flex flex-col gap-3">
                <Skeleton className="h-7 w-2/3 max-w-96" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-11/12" />
                <Skeleton className="h-4 w-4/5" />
                <Skeleton className="mt-5 h-5 w-1/2" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-10/12" />
              </div>
            ) : markdown ? (
              <MessageResponse className="max-w-none text-sm leading-7">{markdown}</MessageResponse>
            ) : (
              <Empty className="min-h-80 border-0">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    {isError ? <AlertCircleIcon /> : <FileTextIcon />}
                  </EmptyMedia>
                  <EmptyTitle>{isError ? "文档读取失败" : "文档没有正文"}</EmptyTitle>
                  <EmptyDescription>
                    {isError
                      ? "暂时无法读取 Markdown 内容，请稍后重试。"
                      : "这个文档当前没有可阅读内容。"}
                  </EmptyDescription>
                </EmptyHeader>
                {isError ? (
                  <EmptyContent>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => void refetch()}
                    >
                      <RefreshCwIcon data-icon="inline-start" />
                      重新加载
                    </Button>
                  </EmptyContent>
                ) : null}
              </Empty>
            )}
          </article>
        </ScrollArea>
      </div>
    </section>
  );
}
