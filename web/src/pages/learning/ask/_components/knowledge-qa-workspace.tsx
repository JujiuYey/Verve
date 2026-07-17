import type { ChatStatus } from "ai";
import { BookOpenIcon, CircleAlertIcon, MessageSquareTextIcon, RotateCcwIcon } from "lucide-react";

import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import { Message, MessageContent, MessageResponse } from "@/components/ai-elements/message";
import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputTextarea,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import { Source, Sources, SourcesContent, SourcesTrigger } from "@/components/ai-elements/sources";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";

import type {
  KnowledgeQAAssistantMessage,
  KnowledgeQAConversationMessage,
} from "../_lib/conversation-state";

export interface KnowledgeQAWorkspaceProps {
  messages: KnowledgeQAConversationMessage[];
  busy: boolean;
  onSend: (question: string) => void;
  onRetry: (assistantId: string) => void;
}

export function KnowledgeQAWorkspace({
  messages,
  busy,
  onSend,
  onRetry,
}: KnowledgeQAWorkspaceProps) {
  const handleSubmit = (message: PromptInputMessage) => {
    if (busy || !message.text.trim()) return;
    onSend(message.text);
  };

  const chatStatus: ChatStatus = busy
    ? messages.some((message) => message.role === "assistant" && message.status === "generating")
      ? "streaming"
      : "submitted"
    : "ready";

  return (
    <main className="flex h-full min-h-0 flex-col overflow-hidden">
      <header className="border-b px-4 py-3 sm:px-6">
        <h1 className="text-lg font-semibold">知识问答</h1>
      </header>

      <Conversation className="min-h-0 flex-1">
        <ConversationContent className="mx-auto min-h-full w-full max-w-3xl gap-7 px-4 py-6 sm:px-6">
          {messages.length === 0 ? (
            <Empty className="min-h-[50vh] border-0 p-6">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquareTextIcon />
                </EmptyMedia>
                <EmptyTitle>从一个问题开始</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : (
            messages.map((message) => (
              <Message from={message.role} key={message.id}>
                <MessageContent className={message.role === "assistant" ? "w-full" : undefined}>
                  {message.role === "user" ? (
                    <div className="whitespace-pre-wrap text-sm leading-6">{message.content}</div>
                  ) : (
                    <AssistantMessage message={message} onRetry={onRetry} />
                  )}
                </MessageContent>
              </Message>
            ))
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      <div className="border-t bg-background px-4 py-3 sm:px-6">
        <PromptInput className="mx-auto max-w-3xl" onSubmit={handleSubmit}>
          <PromptInputBody>
            <PromptInputTextarea
              autoFocus
              disabled={busy}
              maxLength={2000}
              placeholder="输入问题..."
            />
          </PromptInputBody>
          <PromptInputFooter className="justify-end">
            <PromptInputSubmit disabled={busy} status={chatStatus} />
          </PromptInputFooter>
        </PromptInput>
      </div>
    </main>
  );
}

function AssistantMessage({
  message,
  onRetry,
}: {
  message: KnowledgeQAAssistantMessage;
  onRetry: (assistantId: string) => void;
}) {
  if (message.status === "retrieving" || message.status === "generating") {
    return (
      <div aria-live="polite" className="flex items-center gap-2 text-sm text-muted-foreground">
        <Spinner />
        <span>{message.status === "retrieving" ? "正在检索 Wiki..." : "正在整理回答..."}</span>
      </div>
    );
  }

  if (message.status === "failed") {
    return (
      <Alert variant="destructive">
        <CircleAlertIcon />
        <AlertTitle>回答失败</AlertTitle>
        <AlertDescription>
          <p>{message.error || "问答请求失败，请重试"}</p>
          <Button size="sm" type="button" variant="outline" onClick={() => onRetry(message.id)}>
            <RotateCcwIcon data-icon="inline-start" />
            重试
          </Button>
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <article className="flex w-full min-w-0 flex-col gap-5">
      <section aria-labelledby={`${message.id}-knowledge`} className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold" id={`${message.id}-knowledge`}>
          知识回答
        </h2>
        <MessageResponse className="max-w-none text-sm leading-7">
          {message.knowledgeAnswer}
        </MessageResponse>
      </section>
      <Separator />
      <section aria-labelledby={`${message.id}-advice`} className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold" id={`${message.id}-advice`}>
          学习建议
        </h2>
        <MessageResponse className="max-w-none text-sm leading-7">
          {message.learningAdvice}
        </MessageResponse>
      </section>
      {message.sources.length > 0 ? (
        <Sources>
          <SourcesTrigger count={message.sources.length} />
          <SourcesContent className="w-full gap-3">
            {message.sources.map((source) => (
              <Source
                className="items-start text-muted-foreground"
                key={`${source.documentId}:${source.headingPath}`}
                title={source.documentTitle}
              >
                <BookOpenIcon className="mt-0.5 size-4 shrink-0" />
                <span className="flex min-w-0 flex-1 flex-col gap-0.5">
                  <span className="font-medium text-foreground break-words">
                    {source.documentTitle}
                  </span>
                  <span className="break-words">
                    {[source.folderPath, source.headingPath].filter(Boolean).join(" / ") ||
                      "文档正文"}
                  </span>
                </span>
                <span className="shrink-0 tabular-nums">{Math.round(source.score * 100)}%</span>
              </Source>
            ))}
          </SourcesContent>
        </Sources>
      ) : null}
    </article>
  );
}
