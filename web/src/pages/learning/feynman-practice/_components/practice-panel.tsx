import {
  CircleAlertIcon,
  CircleCheckIcon,
  CircleHelpIcon,
  EarIcon,
  MessageSquareTextIcon,
  SendIcon,
  SquareIcon,
} from "lucide-react";

import type { LearningExplanationReview } from "@/api/learning";
import { MessageResponse } from "@/components/ai-elements/message";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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

import { FeynmanAnswerEditor } from "./feynman-answer-editor";

type PracticePanelProps = {
  answer: string;
  turns: LearningExplanationReview[];
  disabled: boolean;
  isSubmitting: boolean;
  isCompleting: boolean;
  isCompleted: boolean;
  completedSummary: string;
  onAnswerChange: (value: string) => void;
  onSubmit: () => void;
  onComplete: () => void;
};

export function PracticePanel({
  answer,
  turns,
  disabled,
  isSubmitting,
  isCompleting,
  isCompleted,
  completedSummary,
  onAnswerChange,
  onSubmit,
  onComplete,
}: PracticePanelProps) {
  const latestTurn = turns.at(-1);
  const readyToWrapUp = latestTurn?.ready_to_wrap_up === true;

  return (
    <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border bg-background">
      <div className="flex shrink-0 flex-col gap-1 border-b px-4 py-3">
        <div className="text-sm font-medium">把整篇文章讲给第一次接触它的人</div>
        <p className="text-xs leading-5 text-muted-foreground">
          用你自己的顺序讲清楚即可。可以一次讲完，也可以根据追问继续补充；代码只在你认为有帮助时使用。
        </p>
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-5 p-4">
          {turns.length > 0 ? (
            <div className="flex flex-col gap-5">
              {turns.map((turn, index) => (
                <ExplanationTurn key={turn.id} turn={turn} index={index} />
              ))}
            </div>
          ) : (
            <Empty className="min-h-48 border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquareTextIcon />
                </EmptyMedia>
                <EmptyTitle>从你的理解开始</EmptyTitle>
                <EmptyDescription>
                  不需要背原文。试着说明文章在解决什么问题、各部分怎样连起来，以及你会怎么讲给别人。
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}

          {isCompleted ? (
            <Alert>
              <CircleCheckIcon />
              <AlertTitle>本次练习已结束</AlertTitle>
              <AlertDescription>
                <p>{completedSummary || "你的解释记录已经保存。"}</p>
              </AlertDescription>
            </Alert>
          ) : (
            <div className="flex flex-col gap-3">
              {readyToWrapUp ? (
                <Alert>
                  <CircleCheckIcon />
                  <AlertTitle>这次解释已经连成完整脉络</AlertTitle>
                  <AlertDescription>
                    <p>如果你没有新的补充，可以在这里结束本次练习；想再换一种说法也可以继续。</p>
                  </AlertDescription>
                </Alert>
              ) : latestTurn?.follow_up_question ? (
                <Alert>
                  <CircleHelpIcon />
                  <AlertTitle>接着讲这一点</AlertTitle>
                  <AlertDescription>
                    <p>{latestTurn.follow_up_question}</p>
                  </AlertDescription>
                </Alert>
              ) : null}

              <FeynmanAnswerEditor
                value={answer}
                onChange={onAnswerChange}
                disabled={disabled || isSubmitting || isCompleting || isCompleted}
                placeholder={
                  turns.length > 0
                    ? "继续补充你的解释，回应上面的疑问，或者换一种更通俗的说法。"
                    : "假设对方完全没读过这篇文章，从头用自己的话讲给他听。"
                }
              />

              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="text-xs text-muted-foreground">
                  {turns.length > 0 ? `已经完成 ${turns.length} 轮解释` : "第一轮可以先讲整体脉络"}
                </p>
                <div className="flex flex-wrap gap-2">
                  {turns.length > 0 && !readyToWrapUp ? (
                    <Button
                      variant="outline"
                      onClick={onComplete}
                      disabled={isCompleting || isSubmitting}
                    >
                      {isCompleting ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <SquareIcon data-icon="inline-start" />
                      )}
                      结束本次练习
                    </Button>
                  ) : null}
                  <Button
                    variant={readyToWrapUp ? "outline" : "default"}
                    onClick={onSubmit}
                    disabled={disabled || !answer.trim()}
                  >
                    {isSubmitting ? (
                      <Spinner data-icon="inline-start" />
                    ) : (
                      <SendIcon data-icon="inline-start" />
                    )}
                    {isSubmitting ? "正在听你的解释" : turns.length > 0 ? "继续补充" : "提交解释"}
                  </Button>
                  {readyToWrapUp ? (
                    <Button onClick={onComplete} disabled={isCompleting || isSubmitting}>
                      {isCompleting ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <SquareIcon data-icon="inline-start" />
                      )}
                      结束本次练习
                    </Button>
                  ) : null}
                </div>
              </div>
            </div>
          )}
        </div>
      </ScrollArea>
    </section>
  );
}

function ExplanationTurn({ turn, index }: { turn: LearningExplanationReview; index: number }) {
  return (
    <article className="flex flex-col gap-4">
      {index > 0 ? <Separator /> : null}
      <div className="flex items-center justify-between gap-3">
        <Badge variant="outline">第 {index + 1} 轮</Badge>
        <time className="text-xs text-muted-foreground" dateTime={turn.created_at}>
          {formatTurnTime(turn.created_at)}
        </time>
      </div>

      <div className="flex flex-col gap-2 rounded-lg bg-muted/30 p-4">
        <div className="text-xs font-medium text-muted-foreground">你的解释</div>
        <MessageResponse className="max-w-none text-sm leading-7">
          {turn.explanation}
        </MessageResponse>
      </div>

      <div className="flex flex-col gap-4 px-1">
        <FeedbackBlock icon={EarIcon} title="我听到的是">
          <p>{turn.heard_summary}</p>
        </FeedbackBlock>

        {turn.clear_points.length > 0 ? (
          <FeedbackList icon={CircleCheckIcon} title="已经讲清楚的地方" items={turn.clear_points} />
        ) : null}

        {turn.confusing_points.length > 0 ? (
          <FeedbackList
            icon={CircleHelpIcon}
            title="我还没听明白的地方"
            items={turn.confusing_points}
          />
        ) : null}

        {turn.misconceptions.length > 0 ? (
          <FeedbackList
            icon={CircleAlertIcon}
            title="可能需要重新确认的理解"
            items={turn.misconceptions}
          />
        ) : null}

        {!turn.context_sufficient ? (
          <Alert>
            <CircleAlertIcon />
            <AlertTitle>文章依据还不够</AlertTitle>
            <AlertDescription>
              <p>当前检索到的原文不足以核对这一部分，反馈只指出疑问，不会把它当成事实错误。</p>
            </AlertDescription>
          </Alert>
        ) : null}
      </div>
    </article>
  );
}

function FeedbackBlock({
  icon: Icon,
  title,
  children,
}: {
  icon: typeof EarIcon;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-[20px_minmax(0,1fr)] gap-2 text-sm leading-6">
      <Icon className="mt-1 size-4 text-muted-foreground" />
      <div className="flex flex-col gap-1">
        <div className="font-medium">{title}</div>
        <div className="text-muted-foreground">{children}</div>
      </div>
    </div>
  );
}

function FeedbackList({
  icon,
  title,
  items,
}: {
  icon: typeof EarIcon;
  title: string;
  items: string[];
}) {
  return (
    <FeedbackBlock icon={icon} title={title}>
      <ul className="flex list-disc flex-col gap-1 pl-5">
        {items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </FeedbackBlock>
  );
}

function formatTurnTime(value: string) {
  if (!value) return "";
  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
