import { IconRefresh } from "@tabler/icons-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import type { QueueStats, TaskInfo } from "@/api/system/queue";
import { queueApi } from "@/api/system/queue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import { StatsCards } from "./_components/stats-cards";
import { TaskTabs } from "./_components/task-tabs";

export function QueuePage() {
  const [stats, setStats] = useState<QueueStats | null>(null);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentState, setCurrentState] = useState("pending");
  const [currentPage, setCurrentPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const refreshIntervalRef = useRef<number | null>(null);

  const loadStats = useCallback(async () => {
    try {
      const data = await queueApi.getStats();
      setStats(data);
    } catch (error: any) {
      toast.error("加载统计失败", {
        description: error.response?.data?.error || "获取队列统计失败",
      });
    }
  }, []);

  const loadTasks = useCallback(async () => {
    setLoading(true);
    try {
      const response = await queueApi.listTasks("default", currentState, pageSize, currentPage);
      setTasks(response.tasks || []);
      setTotal(response.total);
    } catch (error: any) {
      toast.error("加载任务失败", {
        description: error.response?.data?.error || "获取任务列表失败",
      });
    } finally {
      setLoading(false);
    }
  }, [currentState, pageSize, currentPage]);

  const refreshData = useCallback(async () => {
    await Promise.all([loadStats(), loadTasks()]);
  }, [loadStats, loadTasks]);

  // 初始加载和状态/分页变化时重新加载
  useEffect(() => {
    refreshData();
  }, [refreshData]);

  // 自动刷新
  useEffect(() => {
    if (autoRefresh) {
      refreshIntervalRef.current = window.setInterval(() => {
        refreshData();
      }, 5000);
    } else {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
    }
    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
      }
    };
  }, [autoRefresh, refreshData]);

  // 切换状态时重置页码
  const handleStateChange = (state: string) => {
    setCurrentState(state);
    setCurrentPage(0);
  };

  // 页大小变化时重置页码
  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize);
    setCurrentPage(0);
  };

  return (
    <div className="flex h-full w-full flex-col gap-6 p-4 lg:p-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">队列监控</h1>
          <p className="text-muted-foreground text-sm">实时监控任务队列状态和执行情况</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Switch checked={autoRefresh} onCheckedChange={setAutoRefresh} />
            <Label className="text-sm">自动刷新</Label>
          </div>
          <Button size="sm" variant="outline" onClick={() => refreshData()}>
            <IconRefresh className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            刷新
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      <StatsCards stats={stats} />

      {/* 任务列表 */}
      <Card className="flex-1">
        <CardHeader>
          <CardTitle>任务列表</CardTitle>
          <CardDescription>查看不同状态的任务详情</CardDescription>
        </CardHeader>
        <CardContent>
          <TaskTabs
            tasks={tasks}
            loading={loading}
            currentState={currentState}
            page={currentPage}
            pageSize={pageSize}
            total={total}
            onStateChange={handleStateChange}
            onPageChange={setCurrentPage}
            onPageSizeChange={handlePageSizeChange}
          />
        </CardContent>
      </Card>
    </div>
  );
}
