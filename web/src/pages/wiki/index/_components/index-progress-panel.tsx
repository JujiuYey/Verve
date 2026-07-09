import { IconAlertTriangle, IconCheck, IconLoader2 } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { type IndexJobProgress, ragApi } from "@/api/wiki/rag";
import type { Folder } from "@/api/wiki/folder";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

interface IndexBatch {
  rootFolderId: string;
  total: number;
  startedAt: string;
}

interface IndexProgressPanelProps {
  open: boolean;
  folder: Folder | null;
  batch: IndexBatch | null;
  onOpenChange: (open: boolean) => void;
}

export function IndexProgressPanel({
  open,
  folder,
  batch,
  onOpenChange,
}: IndexProgressPanelProps) {
  const [jobs, setJobs] = useState<IndexJobProgress[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open || !folder) return;

    let cancelled = false;
    const loadJobs = async () => {
      setLoading(true);
      try {
        const result = await ragApi.listJobs(folder.id);
        if (!cancelled) setJobs(result || []);
      } catch (error) {
        if (!cancelled) {
          toast.error(error instanceof Error ? error.message : "加载解析进度失败");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    void loadJobs();
    const timer = window.setInterval(() => void loadJobs(), 2000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [folder, open]);

  const visibleJobs = useMemo(() => {
    if (!batch?.startedAt) return jobs;
    const startedAt = new Date(batch.startedAt).getTime();
    return jobs.filter((job) => new Date(job.created_at).getTime() >= startedAt - 1000);
  }, [batch?.startedAt, jobs]);

  const total = batch?.total || visibleJobs.length;
  const running = visibleJobs.filter((job) => job.status === "running").length;
  const completed = visibleJobs.filter((job) => job.status === "completed").length;
  const failed = visibleJobs.filter((job) => job.status === "failed").length;
  const queued = Math.max(total - running - completed - failed, 0);
  const done = completed + failed;
  const progress = total > 0 ? Math.round((done / total) * 100) : 0;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>解析进度</SheetTitle>
          <SheetDescription>{folder ? folder.name : "选择一个知识库根目录查看任务"}</SheetDescription>
        </SheetHeader>

        <div className="flex flex-col gap-4 px-4">
          <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">总体进度</span>
              <span className="font-medium">
                {done}/{total}
              </span>
            </div>
            <Progress value={progress} />
          </div>

          <div className="grid grid-cols-2 gap-2 text-sm">
            <ProgressMetric label="队列中" value={queued} />
            <ProgressMetric label="解析中" value={running} />
            <ProgressMetric label="已完成" value={completed} />
            <ProgressMetric label="失败" value={failed} />
          </div>

          <ScrollArea className="h-[360px] pr-3">
            <div className="flex flex-col gap-2">
              {loading && visibleJobs.length === 0 && (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <IconLoader2 className="h-4 w-4 animate-spin" />
                  正在读取解析任务
                </div>
              )}
              {visibleJobs.length === 0 && !loading && (
                <p className="text-sm text-muted-foreground">暂无解析任务</p>
              )}
              {visibleJobs.map((job) => (
                <div key={job.id} className="rounded-md border p-3">
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate text-sm font-medium">{job.document_id}</span>
                    <JobStatusBadge status={job.status} />
                  </div>
                  <div className="mt-2 text-xs text-muted-foreground">
                    chunks: {job.chunk_count}
                  </div>
                  {job.error_message && (
                    <div className="mt-2 flex gap-2 text-xs text-destructive">
                      <IconAlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                      <span>{job.error_message}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </ScrollArea>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function ProgressMetric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-lg font-semibold">{value}</div>
    </div>
  );
}

function JobStatusBadge({ status }: { status: IndexJobProgress["status"] }) {
  if (status === "completed") {
    return (
      <Badge variant="secondary">
        <IconCheck />
        完成
      </Badge>
    );
  }
  if (status === "failed") {
    return <Badge variant="destructive">失败</Badge>;
  }
  if (status === "running") {
    return (
      <Badge variant="outline">
        <IconLoader2 className="animate-spin" />
        解析中
      </Badge>
    );
  }
  return <Badge variant="outline">队列中</Badge>;
}
