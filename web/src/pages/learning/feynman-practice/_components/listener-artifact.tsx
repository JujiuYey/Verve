import { CircleAlertIcon, CircleCheckIcon, CircleHelpIcon, EarIcon } from "lucide-react";
import type { ReactNode } from "react";

import type { LearningExplanationReview } from "@/api/learning";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function ListenerArtifact({ review }: { review: LearningExplanationReview }) {
  return (
    <div className="flex flex-col gap-3">
      <Feedback icon={EarIcon} title="我听到的是">
        <p>{review.heard_summary}</p>
      </Feedback>
      {review.clear_points.length ? (
        <FeedbackList icon={CircleCheckIcon} title="已经讲清楚" items={review.clear_points} />
      ) : null}
      {review.confusing_points.length ? (
        <FeedbackList icon={CircleHelpIcon} title="还需要说明" items={review.confusing_points} />
      ) : null}
      {review.misconceptions.length ? (
        <FeedbackList icon={CircleAlertIcon} title="需要重新确认" items={review.misconceptions} />
      ) : null}
      {!review.context_sufficient ? (
        <Alert>
          <CircleAlertIcon />
          <AlertTitle>文章依据还不够</AlertTitle>
          <AlertDescription>当前证据不足以核对这一部分。</AlertDescription>
        </Alert>
      ) : null}
      {review.ready_to_wrap_up ? (
        <Alert>
          <CircleCheckIcon />
          <AlertTitle>这次解释已经可以收束</AlertTitle>
          <AlertDescription>没有新的补充时，可以结束本次练习。</AlertDescription>
        </Alert>
      ) : review.follow_up_question ? (
        <Alert>
          <CircleHelpIcon />
          <AlertTitle>接着讲这一点</AlertTitle>
          <AlertDescription>{review.follow_up_question}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}

function Feedback({
  icon: Icon,
  title,
  children,
}: {
  icon: typeof EarIcon;
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="grid grid-cols-[20px_minmax(0,1fr)] gap-2 text-sm leading-6">
      <Icon className="mt-1 size-4 text-muted-foreground" />
      <div>
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
    <Feedback icon={icon} title={title}>
      <ul className="flex list-disc flex-col gap-1 pl-5">
        {items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </Feedback>
  );
}
