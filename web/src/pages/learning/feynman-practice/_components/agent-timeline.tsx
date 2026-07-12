import { EarIcon, FilePenIcon, GraduationCapIcon, MessageSquareTextIcon } from "lucide-react";

import type { LearningAgentType, TimelineItem } from "@/api/learning";
import { Message, MessageContent, MessageResponse } from "@/components/ai-elements/message";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";

import { CuratorArtifact } from "./curator-artifact";
import { ListenerArtifact } from "./listener-artifact";
import { TeacherArtifact } from "./teacher-artifact";

const agentMeta: Record<LearningAgentType, { label: string; icon: typeof EarIcon }> = {
  listener: { label: "倾听", icon: EarIcon },
  teacher: { label: "讲解", icon: GraduationCapIcon },
  curator: { label: "修订", icon: FilePenIcon },
};

const emptyCopy: Record<LearningAgentType, { title: string; description: string }> = {
  listener: {
    title: "开始第一次讲解",
    description: "用自己的话解释文章，倾听者会指出已经讲清楚和仍需澄清的部分。",
  },
  teacher: {
    title: "提出需要补充的问题",
    description: "写下卡住你的概念或关系，老师会结合当前文档进行讲解。",
  },
  curator: {
    title: "提出文档修订目标",
    description: "说明希望补充、纠正或重写的内容，再确认生成的修订建议。",
  },
};

export function AgentTimeline({
  agentType,
  items,
  busyChangeRequestId,
  onApply,
  onCancel,
  onRegenerate,
}: {
  agentType: LearningAgentType;
  items: TimelineItem[];
  busyChangeRequestId: string;
  onApply: (item: TimelineItem) => void;
  onCancel: (item: TimelineItem) => void;
  onRegenerate: (item: TimelineItem) => void;
}) {
  const empty = emptyCopy[agentType];
  return (
    <ScrollArea className="min-h-0 flex-1">
      <div className="flex flex-col gap-5 p-4">
        {items.length ? (
          items.map((item, index) => (
            <TimelineTurn
              key={item.turn.id}
              item={item}
              first={index === 0}
              busyChangeRequestId={busyChangeRequestId}
              onApply={onApply}
              onCancel={onCancel}
              onRegenerate={onRegenerate}
            />
          ))
        ) : (
          <Empty className="min-h-48 border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <MessageSquareTextIcon />
              </EmptyMedia>
              <EmptyTitle>{empty.title}</EmptyTitle>
              <EmptyDescription>{empty.description}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </div>
    </ScrollArea>
  );
}

function TimelineTurn({
  item,
  first,
  busyChangeRequestId,
  onApply,
  onCancel,
  onRegenerate,
}: {
  item: TimelineItem;
  first: boolean;
  busyChangeRequestId: string;
  onApply: (item: TimelineItem) => void;
  onCancel: (item: TimelineItem) => void;
  onRegenerate: (item: TimelineItem) => void;
}) {
  const meta = agentMeta[item.turn.agent_type];
  const Icon = meta.icon;
  return (
    <article className="flex min-h-28 flex-col gap-3">
      {!first ? <Separator /> : null}
      <div className="flex items-center justify-between gap-3">
        <Badge variant="outline">
          <Icon />
          {meta.label}
        </Badge>
        <time className="text-xs text-muted-foreground" dateTime={item.turn.created_at}>
          {formatTime(item.turn.created_at)}
        </time>
      </div>
      <Message from="user">
        <MessageContent>
          <MessageResponse>{item.user_message.content}</MessageResponse>
        </MessageContent>
      </Message>
      {item.turn.status === "processing" ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Spinner />
          正在处理
        </div>
      ) : null}
      {item.turn.status === "failed" ? (
        <Alert>
          <AlertTitle>这一轮处理失败</AlertTitle>
          <AlertDescription>{item.turn.error_message || "请重新提交这一轮。"}</AlertDescription>
        </Alert>
      ) : null}
      {item.assistant_message && item.turn.agent_type !== "listener" ? (
        <Message from="assistant">
          <MessageContent>
            <MessageResponse>{item.assistant_message.content}</MessageResponse>
          </MessageContent>
        </Message>
      ) : null}
      {item.artifact?.type === "explanation_review" ? (
        <ListenerArtifact review={item.artifact.data} />
      ) : null}
      {item.artifact?.type === "teaching_intervention" ? (
        <TeacherArtifact intervention={item.artifact.data} />
      ) : null}
      {item.artifact?.type === "wiki_change_request" ? (
        <CuratorArtifact
          request={item.artifact.data}
          busy={busyChangeRequestId === item.artifact.data.id}
          onApply={() => onApply(item)}
          onCancel={() => onCancel(item)}
          onRegenerate={() => onRegenerate(item)}
        />
      ) : null}
    </article>
  );
}

function formatTime(value: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", { hour: "2-digit", minute: "2-digit" }).format(
        new Date(value),
      )
    : "";
}
