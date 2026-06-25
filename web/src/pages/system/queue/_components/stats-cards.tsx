import {
  IconActivity,
  IconAlertCircle,
  IconArchive,
  IconCalendarEvent,
  IconCircleCheck,
  IconClock,
} from "@tabler/icons-react";

import type { QueueStats } from "@/api/system/queue";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface StatsCardsProps {
  stats: QueueStats | null;
}

const statsConfig = [
  {
    title: "等待处理",
    key: "pending" as const,
    icon: IconClock,
    color: "text-blue-500",
  },
  {
    title: "正在处理",
    key: "active" as const,
    icon: IconActivity,
    color: "text-green-500",
  },
  {
    title: "计划执行",
    key: "scheduled" as const,
    icon: IconCalendarEvent,
    color: "text-purple-500",
  },
  {
    title: "重试中",
    key: "retry" as const,
    icon: IconAlertCircle,
    color: "text-orange-500",
  },
  {
    title: "已归档",
    key: "archived" as const,
    icon: IconArchive,
    color: "text-gray-500",
  },
  {
    title: "已完成",
    key: "completed" as const,
    icon: IconCircleCheck,
    color: "text-emerald-500",
  },
];

export function StatsCards({ stats }: StatsCardsProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
      {statsConfig.map((card) => {
        const Icon = card.icon;
        return (
          <Card key={card.key}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
              <Icon className={`h-4 w-4 ${card.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats?.[card.key] ?? 0}</div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
