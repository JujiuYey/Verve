import {
  BookOpenIcon,
  BotIcon,
  CalendarDaysIcon,
  CheckCircle2Icon,
  Clock3Icon,
  FilePenLineIcon,
  FlameIcon,
  MessageSquareTextIcon,
  RouteIcon,
  SparklesIcon,
  TargetIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import { SagPage } from "@/components/sag-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

type PracticeLog = {
  id: string;
  date: string;
  time: string;
  document: string;
  folder: string;
  status: "completed" | "active";
  duration: number;
  turns: number;
  agents: string[];
  summary: string;
  learned: string[];
  unresolved: string[];
  corrections: string[];
  nextActions: string[];
  documentChanges: string[];
  timeline: { agent: string; action: string }[];
};

type ContributionDay = {
  date: string;
  count: number;
};

const PRACTICE_LOGS: PracticeLog[] = [
  {
    id: "practice-go-values",
    date: "2026-07-11",
    time: "21:10",
    document: "Go 的值、类型与变量",
    folder: "Go 语言基础",
    status: "completed",
    duration: 38,
    turns: 4,
    agents: ["FeynmanListener", "LearningTeacher", "WikiCurator"],
    summary:
      "从“值拥有具体类型”出发，逐步讲清变量只是保存值的名字，并补齐了静态类型与运行时值之间的关系。",
    learned: [
      "能够用自己的语言解释值、类型和变量三者的关系",
      "能够说明类型信息为什么会影响可执行的操作",
      "能够区分变量声明与变量当前保存的值",
    ],
    unresolved: ["接口值同时包含动态类型和动态值的部分仍需单独练习"],
    corrections: ["“变量有类型”需要进一步表述为变量声明约束了它能够保存的值"],
    nextActions: ["阅读《接口值的内部结构》", "用一个 nil 接口示例继续费曼讲解"],
    documentChanges: ["补充“变量不是盒子类型本身”的解释", "增加接口值章节的前置链接"],
    timeline: [
      { agent: "LearningSupervisor", action: "根据最近记忆选择本文作为练习对象" },
      { agent: "FeynmanListener", action: "倾听 4 轮解释并识别一个表述歧义" },
      { agent: "LearningTeacher", action: "补充静态类型与运行时值的区别" },
      { agent: "WikiCurator", action: "发现原文缺少变量语义说明并提出两处修改" },
    ],
  },
  {
    id: "practice-interface",
    date: "2026-07-10",
    time: "20:35",
    document: "接口与隐式实现",
    folder: "Go 语言基础",
    status: "active",
    duration: 24,
    turns: 3,
    agents: ["FeynmanListener", "LearningTeacher"],
    summary: "已经能解释接口不要求显式声明，但对方法集与指针接收者的关系仍不稳定。",
    learned: ["理解接口通过方法集建立约束", "能够解释隐式实现降低了类型之间的耦合"],
    unresolved: ["值方法与指针方法分别属于哪个方法集", "为什么 *T 可以使用值接收者方法"],
    corrections: ["实现接口的不是“文件”或“包”，而是具体类型的方法集"],
    nextActions: ["先由 LearningTeacher 演示方法集表格", "完成一轮不依赖代码的口头解释"],
    documentChanges: [],
    timeline: [
      { agent: "LearningSupervisor", action: "恢复上次未结束的接口练习" },
      { agent: "FeynmanListener", action: "确认隐式实现部分已经讲清" },
      { agent: "LearningTeacher", action: "正在补充方法集与接收者规则" },
    ],
  },
  {
    id: "practice-errors",
    date: "2026-07-08",
    time: "19:50",
    document: "错误处理与 errors.Is",
    folder: "Go 工程化",
    status: "completed",
    duration: 46,
    turns: 5,
    agents: ["FeynmanListener", "LearningTeacher"],
    summary: "围绕错误包装完成了五轮解释，最终能够说明 errors.Is 为什么沿错误链进行匹配。",
    learned: ["理解 %w 会保留错误链", "能够区分 errors.Is 与字符串比较"],
    unresolved: ["自定义 Is 方法的匹配边界还没有形成稳定理解"],
    corrections: ["errors.Is 并不是比较两个错误字符串是否相同"],
    nextActions: ["阅读自定义错误类型", "在下一次练习中解释 errors.As"],
    documentChanges: [],
    timeline: [
      { agent: "LearningSupervisor", action: "根据上次薄弱点安排错误链练习" },
      { agent: "FeynmanListener", action: "完成五轮倾听与追问" },
      { agent: "LearningTeacher", action: "使用错误链示意补充解释" },
    ],
  },
];

const CONTRIBUTION_OVERRIDES: Record<string, number> = {
  "2026-07-11": 2,
  "2026-07-10": 1,
  "2026-07-08": 1,
  "2026-07-06": 2,
  "2026-07-04": 1,
  "2026-07-03": 3,
  "2026-06-29": 1,
  "2026-06-27": 2,
  "2026-06-24": 1,
  "2026-06-20": 2,
  "2026-06-18": 1,
};

const CONTRIBUTION_DAYS = buildContributionDays();

export function JournalPage() {
  const [selectedLogID, setSelectedLogID] = useState(PRACTICE_LOGS[0].id);
  const [selectedDate, setSelectedDate] = useState("2026-07-11");
  const selectedLog = PRACTICE_LOGS.find((item) => item.id === selectedLogID) ?? PRACTICE_LOGS[0];
  const selectedActivity = useMemo(
    () => CONTRIBUTION_DAYS.find((item) => item.date === selectedDate),
    [selectedDate],
  );

  return (
    <SagPage
      title={
        <span className="flex flex-wrap items-center gap-3">
          <span>练习日志</span>
          <Badge variant="secondary">Mock 数据</Badge>
        </span>
      }
      description="查看有效练习、解释变化和 Agent 协作记录。当前页面用于反推练习日志的数据结构。"
      bodyClassName="overflow-hidden"
    >
      <ScrollArea className="h-full">
        <div className="mx-auto flex w-full max-w-[1480px] flex-col gap-6 pb-6 pr-3">
          <section className="grid border-y sm:grid-cols-2 xl:grid-cols-4">
            <Metric label="本周有效练习" value="7" suffix="次" icon={TargetIcon} />
            <Metric label="连续练习" value="6" suffix="天" icon={FlameIcon} />
            <Metric label="完成文档" value="12" suffix="篇" icon={BookOpenIcon} />
            <Metric label="累计解释" value="37" suffix="轮" icon={MessageSquareTextIcon} />
          </section>

          <section className="rounded-lg border bg-background">
            <div className="flex flex-col gap-2 border-b px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 className="font-semibold">有效练习贡献</h2>
                <p className="mt-1 text-xs text-muted-foreground">
                  颜色深浅表示当天完成的有效练习会话数
                </p>
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>少</span>
                {[0, 1, 2, 3, 4].map((level) => (
                  <span
                    key={level}
                    className={cn("size-3 rounded-[2px]", contributionClass(level))}
                  />
                ))}
                <span>多</span>
              </div>
            </div>
            <div className="overflow-x-auto px-5 py-5">
              <div className="min-w-[720px]">
                <div className="mb-2 pl-10 text-xs text-muted-foreground">最近 26 周</div>
                <div className="flex gap-2">
                  <div className="grid h-[122px] grid-rows-7 gap-1 text-[10px] text-muted-foreground">
                    <span />
                    <span>一</span>
                    <span />
                    <span>三</span>
                    <span />
                    <span>五</span>
                    <span />
                  </div>
                  <div className="grid grid-flow-col grid-rows-7 gap-1">
                    {CONTRIBUTION_DAYS.map((day) => (
                      <button
                        key={day.date}
                        type="button"
                        aria-label={`${day.date}，${day.count} 次有效练习`}
                        title={`${day.date} · ${day.count} 次有效练习`}
                        className={cn(
                          "size-3.5 rounded-[2px] outline-none ring-offset-2 transition-transform hover:scale-125 focus-visible:ring-2 focus-visible:ring-ring",
                          contributionClass(day.count),
                          selectedDate === day.date && "ring-2 ring-foreground ring-offset-1",
                        )}
                        onClick={() => setSelectedDate(day.date)}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2 border-t px-5 py-3 text-sm">
              <CalendarDaysIcon className="size-4 text-muted-foreground" />
              <span>{formatDate(selectedDate)}</span>
              <span className="text-muted-foreground">
                {selectedActivity?.count ?? 0} 次有效练习
              </span>
            </div>
          </section>

          <div className="grid min-h-[620px] gap-5 xl:grid-cols-[420px_minmax(0,1fr)]">
            <section className="overflow-hidden rounded-lg border bg-background">
              <div className="border-b px-4 py-3">
                <h2 className="font-semibold">最近练习</h2>
                <p className="mt-1 text-xs text-muted-foreground">选择一条记录查看完整学习过程</p>
              </div>
              <div className="divide-y">
                {PRACTICE_LOGS.map((log) => (
                  <button
                    key={log.id}
                    type="button"
                    className={cn(
                      "flex w-full flex-col gap-3 px-4 py-4 text-left transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring",
                      selectedLogID === log.id && "bg-muted/60",
                    )}
                    onClick={() => setSelectedLogID(log.id)}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="truncate font-medium">{log.document}</div>
                        <div className="mt-1 truncate text-xs text-muted-foreground">
                          {log.folder}
                        </div>
                      </div>
                      <StatusBadge status={log.status} />
                    </div>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
                      <span>{formatDate(log.date)}</span>
                      <span className="inline-flex items-center gap-1">
                        <Clock3Icon className="size-3.5" />
                        {log.duration} 分钟
                      </span>
                      <span className="inline-flex items-center gap-1">
                        <MessageSquareTextIcon className="size-3.5" />
                        {log.turns} 轮解释
                      </span>
                    </div>
                    <div className="flex flex-wrap gap-1.5">
                      {log.agents.map((agent) => (
                        <Badge key={agent} variant="outline" className="font-normal">
                          {agent}
                        </Badge>
                      ))}
                    </div>
                  </button>
                ))}
              </div>
            </section>

            <PracticeDetail log={selectedLog} />
          </div>
        </div>
      </ScrollArea>
    </SagPage>
  );
}

function Metric({
  label,
  value,
  suffix,
  icon: Icon,
}: {
  label: string;
  value: string;
  suffix: string;
  icon: typeof TargetIcon;
}) {
  return (
    <div className="flex min-h-28 items-center gap-4 border-b p-5 last:border-b-0 sm:odd:border-r sm:[&:nth-last-child(-n+2)]:border-b-0 xl:border-b-0 xl:border-r xl:last:border-r-0">
      <div className="flex size-10 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
        <Icon className="size-5" />
      </div>
      <div>
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className="mt-1 flex items-baseline gap-1">
          <span className="text-2xl font-semibold tabular-nums">{value}</span>
          <span className="text-sm text-muted-foreground">{suffix}</span>
        </div>
      </div>
    </div>
  );
}

function PracticeDetail({ log }: { log: PracticeLog }) {
  return (
    <section className="rounded-lg border bg-background">
      <div className="flex flex-col gap-3 border-b px-5 py-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-lg font-semibold">{log.document}</h2>
            <StatusBadge status={log.status} />
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {formatDate(log.date)} {log.time} · {log.folder} · {log.duration} 分钟
          </p>
        </div>
        <Button variant="outline" size="sm">
          打开练习
        </Button>
      </div>

      <div className="flex flex-col gap-6 p-5">
        <div>
          <SectionTitle icon={SparklesIcon}>本次练习总结</SectionTitle>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">{log.summary}</p>
        </div>

        <Separator />

        <div className="grid gap-6 lg:grid-cols-2">
          <DetailList icon={CheckCircle2Icon} title="已经讲清" items={log.learned} />
          <DetailList icon={TargetIcon} title="仍需继续" items={log.unresolved} />
          <DetailList icon={MessageSquareTextIcon} title="关键纠正" items={log.corrections} />
          <DetailList icon={RouteIcon} title="下一步" items={log.nextActions} />
        </div>

        <Separator />

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(280px,0.8fr)]">
          <div>
            <SectionTitle icon={BotIcon}>Agent 协作记录</SectionTitle>
            <ol className="mt-4 flex flex-col">
              {log.timeline.map((item, index) => (
                <li
                  key={`${item.agent}-${item.action}`}
                  className="grid grid-cols-[20px_minmax(0,1fr)] gap-3"
                >
                  <div className="flex flex-col items-center">
                    <span className="mt-1 size-2.5 rounded-full bg-foreground" />
                    {index < log.timeline.length - 1 ? (
                      <span className="my-1 w-px flex-1 bg-border" />
                    ) : null}
                  </div>
                  <div className="pb-4">
                    <div className="text-sm font-medium">{item.agent}</div>
                    <p className="mt-1 text-sm leading-6 text-muted-foreground">{item.action}</p>
                  </div>
                </li>
              ))}
            </ol>
          </div>

          <div>
            <SectionTitle icon={FilePenLineIcon}>Wiki 文档变化</SectionTitle>
            {log.documentChanges.length > 0 ? (
              <ul className="mt-4 flex list-disc flex-col gap-2 pl-5 text-sm leading-6 text-muted-foreground">
                {log.documentChanges.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            ) : (
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                本次练习没有修改 Wiki 文档。
              </p>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

function DetailList({
  icon: Icon,
  title,
  items,
}: {
  icon: typeof TargetIcon;
  title: string;
  items: string[];
}) {
  return (
    <div>
      <SectionTitle icon={Icon}>{title}</SectionTitle>
      <ul className="mt-3 flex list-disc flex-col gap-2 pl-5 text-sm leading-6 text-muted-foreground">
        {items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </div>
  );
}

function SectionTitle({
  icon: Icon,
  children,
}: {
  icon: typeof TargetIcon;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center gap-2 text-sm font-semibold">
      <Icon className="size-4 text-muted-foreground" />
      <span>{children}</span>
    </div>
  );
}

function StatusBadge({ status }: { status: PracticeLog["status"] }) {
  return status === "completed" ? (
    <Badge variant="secondary">已完成</Badge>
  ) : (
    <Badge variant="outline">进行中</Badge>
  );
}

function contributionClass(count: number) {
  if (count <= 0) return "bg-muted";
  if (count === 1) return "bg-emerald-200 dark:bg-emerald-950";
  if (count === 2) return "bg-emerald-400 dark:bg-emerald-800";
  if (count === 3) return "bg-emerald-600 dark:bg-emerald-600";
  return "bg-emerald-800 dark:bg-emerald-400";
}

function buildContributionDays(): ContributionDay[] {
  const start = Date.UTC(2026, 0, 12);
  return Array.from({ length: 26 * 7 }, (_, index) => {
    const date = new Date(start + index * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
    const generated = index % 19 === 0 ? 2 : index % 11 === 0 ? 1 : index % 37 === 0 ? 3 : 0;
    return { date, count: CONTRIBUTION_OVERRIDES[date] ?? generated };
  });
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "long",
    day: "numeric",
    weekday: "short",
  }).format(new Date(`${value}T00:00:00`));
}
