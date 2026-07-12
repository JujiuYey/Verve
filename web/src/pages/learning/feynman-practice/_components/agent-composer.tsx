import { SendIcon, SquareIcon } from "lucide-react";

import type { LearningAgentType } from "@/api/learning";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { AgentSelector } from "./agent-selector";
import { FeynmanAnswerEditor } from "./feynman-answer-editor";

const copy: Record<LearningAgentType, { placeholder: string; submit: string; progress: string }> = {
  listener: {
    placeholder: "用自己的话解释文章，或者继续回应倾听者的问题。",
    submit: "提交解释",
    progress: "正在听你的解释",
  },
  teacher: {
    placeholder: "写下卡住你的概念、关系或具体问题。",
    submit: "请老师讲解",
    progress: "正在组织讲解",
  },
  curator: {
    placeholder: "说明希望如何补充、纠正或重写当前 Wiki 文档。",
    submit: "生成修订建议",
    progress: "正在生成修订建议",
  },
};

export function AgentComposer({
  selectedAgent,
  value,
  disabled,
  isSubmitting,
  canComplete,
  isCompleting,
  onAgentChange,
  onChange,
  onSubmit,
  onComplete,
}: {
  selectedAgent: LearningAgentType;
  value: string;
  disabled: boolean;
  isSubmitting: boolean;
  canComplete: boolean;
  isCompleting: boolean;
  onAgentChange: (agent: LearningAgentType) => void;
  onChange: (value: string) => void;
  onSubmit: () => void;
  onComplete: () => void;
}) {
  const labels = copy[selectedAgent];
  return (
    <div className="flex shrink-0 flex-col gap-3 border-t bg-background p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <AgentSelector
          value={selectedAgent}
          disabled={disabled || isSubmitting}
          onChange={onAgentChange}
        />
        {canComplete ? (
          <Button
            variant="outline"
            size="sm"
            onClick={onComplete}
            disabled={isCompleting || isSubmitting || !!value.trim()}
          >
            {isCompleting ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <SquareIcon data-icon="inline-start" />
            )}
            结束本次练习
          </Button>
        ) : null}
      </div>
      <FeynmanAnswerEditor
        value={value}
        onChange={onChange}
        disabled={disabled || isSubmitting}
        placeholder={labels.placeholder}
      />
      <div className="flex justify-end">
        <Button onClick={onSubmit} disabled={disabled || isSubmitting || !value.trim()}>
          {isSubmitting ? (
            <Spinner data-icon="inline-start" />
          ) : (
            <SendIcon data-icon="inline-start" />
          )}
          {isSubmitting ? labels.progress : labels.submit}
        </Button>
      </div>
    </div>
  );
}
