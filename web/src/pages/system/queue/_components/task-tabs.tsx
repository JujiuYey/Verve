import type { TaskInfo } from "@/api/system/queue";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { TaskTable } from "./task-table";

interface TaskTabsProps {
  tasks: TaskInfo[];
  loading: boolean;
  currentState: string;
  page: number;
  pageSize: number;
  total: number;
  onStateChange: (state: string) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

const states = [
  { value: "pending", label: "等待中" },
  { value: "active", label: "处理中" },
  { value: "scheduled", label: "已计划" },
  { value: "retry", label: "重试中" },
  { value: "archived", label: "已归档" },
  { value: "completed", label: "已完成" },
];

export function TaskTabs({
  tasks,
  loading,
  currentState,
  page,
  pageSize,
  total,
  onStateChange,
  onPageChange,
  onPageSizeChange,
}: TaskTabsProps) {
  return (
    <Tabs value={currentState} onValueChange={(value) => onStateChange(value)}>
      <TabsList className="grid w-full grid-cols-6">
        {states.map((s) => (
          <TabsTrigger key={s.value} value={s.value}>
            {s.label}
          </TabsTrigger>
        ))}
      </TabsList>

      {states.map((s) => (
        <TabsContent key={s.value} value={s.value} className="mt-4">
          <TaskTable
            tasks={tasks}
            loading={loading}
            state={s.value}
            page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
          />
        </TabsContent>
      ))}
    </Tabs>
  );
}
