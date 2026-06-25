"use client";

import { Loader2Icon } from "lucide-react";
import { useCallback, useState } from "react";

import {
  Conversation,
  ConversationContent,
  ConversationEmptyState,
  ConversationScrollButton,
} from "../../../../components/ai-elements/conversation";
import {
  Message,
  MessageContent,
  MessageResponse,
  MessageToolbar,
} from "../../../../components/ai-elements/message";
import type { PromptInputMessage } from "../../../../components/ai-elements/prompt-input";
import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputTextarea,
} from "../../../../components/ai-elements/prompt-input";
import { Suggestion, Suggestions } from "../../../../components/ai-elements/suggestion";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolInput,
  ToolOutput,
} from "../../../../components/ai-elements/tool";
import type { AgentMessage, AgentToolCall } from "../_hooks/agent-chat";
import { getToolDisplayName } from "../_hooks/agent-chat";

const WELCOME_CONTENT = `你好！我是 AI 系统助手。基于先进的语言理解能力，我可以用自然对话的方式帮你管理系统。

例如你可以这样说：

• "创建一个技术部，负责系统开发"
• "帮我把研发部的描述改成负责产品研发"
• "删除市场部"
• "有哪些部门？"
• "搜索一下技术相关的部门"

请告诉我你想做什么？`;

const SUGGESTIONS = [
  "创建一个技术部，负责系统开发",
  "帮我把研发部的描述改成负责产品研发",
  "删除市场部",
  "有哪些部门？",
];

export interface ChatPanelProps {
  messages: AgentMessage[];
  loading: boolean;
  onSend: (text: string) => Promise<void>;
}

function ToolCallItem({ toolCall }: { toolCall: AgentToolCall }) {
  const hasInput = toolCall.input && Object.keys(toolCall.input).length > 0;
  const hasOutput = toolCall.output != null;
  const hasError = !!toolCall.error;

  return (
    <Tool>
      <ToolHeader
        title={getToolDisplayName(toolCall.name)}
        type="dynamic-tool"
        state={hasError ? "output-error" : hasOutput ? "output-available" : "input-available"}
        toolName={toolCall.name}
      />
      <ToolContent>
        {hasInput && <ToolInput input={toolCall.input} />}
        <ToolOutput output={toolCall.output} errorText={toolCall.error} />
      </ToolContent>
    </Tool>
  );
}

export function ChatPanel({ messages, loading, onSend }: ChatPanelProps) {
  const [isGenerating, setIsGenerating] = useState(false);

  const handleSubmit = useCallback(
    async (message: PromptInputMessage) => {
      const text = message.text.trim();
      if (!text) return;
      setIsGenerating(true);
      try {
        await onSend(text);
      } finally {
        setIsGenerating(false);
      }
    },
    [onSend],
  );

  const handleStop = useCallback(() => {
    setIsGenerating(false);
  }, []);

  const handleSuggestionClick = useCallback(
    async (suggestion: string) => {
      await onSend(suggestion);
    },
    [onSend],
  );

  return (
    <div className="flex h-full flex-col bg-background">
      {/* 消息列表 */}
      <Conversation>
        <ConversationContent>
          {messages.length === 0 ? (
            <ConversationEmptyState title="你好！我是 AI 系统助手">
              <div className="space-y-4">
                <p className="whitespace-pre-wrap text-sm">{WELCOME_CONTENT}</p>
                <Suggestions>
                  {SUGGESTIONS.map((s) => (
                    <Suggestion key={s} suggestion={s} onClick={handleSuggestionClick} />
                  ))}
                </Suggestions>
              </div>
            </ConversationEmptyState>
          ) : (
            <>
              {messages.map((msg) => (
                <Message key={msg.id} from={msg.role}>
                  <MessageContent>
                    <MessageResponse>{msg.content}</MessageResponse>
                    {msg.tool_calls && msg.tool_calls.length > 0 && (
                      <MessageToolbar>
                        <div className="flex w-full flex-col gap-2">
                          {msg.tool_calls.map((toolCall) => (
                            <ToolCallItem key={toolCall.id} toolCall={toolCall} />
                          ))}
                        </div>
                      </MessageToolbar>
                    )}
                  </MessageContent>
                </Message>
              ))}
              {loading && (
                <Message from="assistant">
                  <MessageContent>
                    <div className="flex items-center gap-2 text-muted-foreground text-sm">
                      <Loader2Icon className="size-4 animate-spin" />
                      <span>思考中...</span>
                    </div>
                  </MessageContent>
                </Message>
              )}
            </>
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      {/* 输入框 */}
      <div className="border-t p-4">
        <PromptInput onSubmit={handleSubmit}>
          <PromptInputBody />
          <PromptInputTextarea placeholder="输入你的指令，例如：创建一个技术部..." />
          <PromptInputFooter className="justify-end">
            <PromptInputSubmit
              onStop={handleStop}
              status={isGenerating ? "streaming" : loading ? "submitted" : undefined}
            />
          </PromptInputFooter>
        </PromptInput>
      </div>
    </div>
  );
}
