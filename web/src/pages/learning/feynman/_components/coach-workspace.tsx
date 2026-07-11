import type { ChatStatus } from "ai";
import { BotIcon, CornerDownLeftIcon, PlayIcon } from "lucide-react";

import type { LearningCoachAction } from "@/api/learning";
import {
  Conversation,
  ConversationContent,
  ConversationEmptyState,
  ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import { Message, MessageContent, MessageResponse } from "@/components/ai-elements/message";
import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputTools,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import { Reasoning, ReasoningContent, ReasoningTrigger } from "@/components/ai-elements/reasoning";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolInput,
  ToolOutput,
} from "@/components/ai-elements/tool";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export type ToolEventState = "input-available" | "output-available" | "output-error";

export type ToolEvent = {
  id: string;
  name: string;
  arguments?: string;
  output?: string;
  state: ToolEventState;
};

export type CoachMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  reasoning?: string;
  toolEvents?: ToolEvent[];
};

export type CoachWorkspaceProps = {
  agentName?: string;
  messages: CoachMessage[];
  status: ChatStatus;
  action: LearningCoachAction | null;
  onSend: (message: string) => void;
  onEnterPractice: () => void;
};

function safeJson(input: string | undefined): unknown {
  if (!input) return "";
  try {
    return JSON.parse(input);
  } catch {
    return input;
  }
}

export function CoachWorkspace({
  agentName,
  messages,
  status,
  action,
  onSend,
  onEnterPractice,
}: CoachWorkspaceProps) {
  const isBusy = status === "submitted" || status === "streaming";

  const handleSubmit = (message: PromptInputMessage) => {
    if (isBusy || !message.text.trim()) return;
    onSend(message.text);
  };

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-hidden p-6">
      <div className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-2xl font-bold">{agentName || "费曼学习 Agent"}</h1>
          <Badge variant="secondary">真实上下文</Badge>
        </div>
        <p className="max-w-3xl text-sm leading-6 text-muted-foreground">
          直接说继续学习。Agent 会查询 Wiki
          文件夹、文档、学习记录和用户画像，再决定下一步进入哪一个小节。
        </p>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border bg-background">
        <div className="flex items-center gap-2 border-b px-4 py-3">
          <BotIcon className="size-4 text-muted-foreground" />
          <div className="text-sm font-medium">学习调度</div>
          {isBusy ? <Badge variant="outline">查询中</Badge> : null}
        </div>

        <Conversation className="min-h-0">
          <ConversationContent className="min-h-full gap-6 p-4">
            {messages.length === 0 ? (
              <ConversationEmptyState
                description="它会先读当前资料结构和学习状态，再告诉你该继续、复习，还是先补资料。"
                icon={<BotIcon className="size-8" />}
                title="让 Agent 接管下一步"
              >
                <div className="flex flex-col items-center gap-4">
                  <div className="rounded-full border bg-muted/40 p-3 text-muted-foreground">
                    <BotIcon className="size-6" />
                  </div>
                  <div className="space-y-1">
                    <div className="text-base font-semibold">让 Agent 接管下一步</div>
                    <div className="max-w-md text-sm leading-6 text-muted-foreground">
                      它会先读当前资料结构和学习状态，再告诉你该继续、复习，还是先补资料。
                    </div>
                  </div>
                  <Button onClick={() => onSend("继续学习")} disabled={isBusy}>
                    <PlayIcon className="size-4" />
                    继续学习
                  </Button>
                </div>
              </ConversationEmptyState>
            ) : (
              messages.map((message) => (
                <Message from={message.role} key={message.id}>
                  <MessageContent>
                    {message.role === "assistant" ? (
                      <>
                        {message.reasoning ? (
                          <Reasoning
                            className="w-full"
                            isStreaming={
                              status === "streaming" &&
                              !message.content &&
                              !message.toolEvents?.length
                            }
                          >
                            <ReasoningTrigger />
                            <ReasoningContent>{message.reasoning}</ReasoningContent>
                          </Reasoning>
                        ) : null}

                        {message.toolEvents?.map((tool) => (
                          <Tool defaultOpen key={`${tool.id}-${tool.name}`}>
                            <ToolHeader
                              state={tool.state}
                              title={tool.name}
                              toolName={tool.name}
                              type="dynamic-tool"
                            />
                            <ToolContent>
                              {tool.arguments ? (
                                <ToolInput input={safeJson(tool.arguments)} />
                              ) : null}
                              <ToolOutput
                                errorText={
                                  tool.state === "output-error" ? "Tool failed" : undefined
                                }
                                output={safeJson(tool.output)}
                              />
                            </ToolContent>
                          </Tool>
                        ))}

                        {message.content ? (
                          <MessageResponse className="max-w-none text-sm leading-7">
                            {message.content}
                          </MessageResponse>
                        ) : !message.reasoning && !message.toolEvents?.length ? (
                          <span className="text-sm text-muted-foreground">正在查询上下文...</span>
                        ) : null}
                      </>
                    ) : (
                      <div className="whitespace-pre-wrap text-sm leading-6">{message.content}</div>
                    )}
                  </MessageContent>
                </Message>
              ))
            )}

            {action?.type === "navigate_to_practice" && action.document_id ? (
              <Message from="assistant">
                <MessageContent>
                  <Button onClick={onEnterPractice}>
                    <CornerDownLeftIcon className="size-4" />
                    {action.label || "进入练习"}
                  </Button>
                </MessageContent>
              </Message>
            ) : null}
          </ConversationContent>
          <ConversationScrollButton />
        </Conversation>

        <div className="border-t p-4">
          <PromptInput onSubmit={handleSubmit}>
            <PromptInputBody>
              <PromptInputTextarea
                className="min-h-11"
                disabled={isBusy}
                placeholder="继续学习 / 复习薄弱点 / 今天从哪里开始..."
              />
            </PromptInputBody>
            <PromptInputFooter>
              <PromptInputTools>
                <Badge variant="outline">LearningCoach</Badge>
              </PromptInputTools>
              <PromptInputSubmit disabled={isBusy} status={status} />
            </PromptInputFooter>
          </PromptInput>
        </div>
      </div>
    </div>
  );
}
