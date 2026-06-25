import { IconLoader2, IconSend } from "@tabler/icons-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { chatStream } from "@/api/ai/chat";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Usage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

interface RetrievedChunk {
  chunk_id: number;
  content: string;
  document_id: string;
  filename: string;
  folder_id: string;
}

interface RagMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  retrieved_chunks?: RetrievedChunk[];
  relevance_score?: number;
  usage?: Usage;
  created_at: string;
}

interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
  retrievedChunks?: RetrievedChunk[];
  usage?: Usage;
}

interface ChatPanelProps {
  sessionId?: string;
  folderId?: string;
  onSessionCreated: (sessionId: string) => void;
}

const CONTEXT_LIMIT = 256_000;
const CONTEXT_WARNING_THRESHOLD = 200_000;

const WELCOME_MESSAGE: ChatMessage = {
  id: "welcome",
  role: "assistant",
  content: "你好！我是 RAG 助手，有什么可以帮助你的吗？",
  timestamp: new Date(),
};

function formatTokenCount(tokens: number): string {
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M`;
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}k`;
  return tokens.toString();
}

function generateId(): string {
  return `rag-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

// Mock 会话数据（本地存储，刷新后丢失）
const mockRagSessions: Record<string, RagMessage[]> = {};

export function ChatPanel({ sessionId, folderId, onSessionCreated }: ChatPanelProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([WELCOME_MESSAGE]);
  const [isLoading, setIsLoading] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [sessionTotalTokens, setSessionTotalTokens] = useState<number | undefined>();
  const currentSessionIdRef = useRef<string | undefined>(sessionId);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  // 用于流式更新时追踪当前 AI 消息
  const currentAssistantMessageRef = useRef<string | null>(null);

  const scrollToBottom = useCallback(() => {
    requestAnimationFrame(() => {
      if (scrollRef.current) {
        scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
      }
    });
  }, []);

  // Load history when sessionId changes

  useEffect(() => {
    currentSessionIdRef.current = sessionId;

    if (sessionId) {
      const history = mockRagSessions[sessionId];
      if (history && history.length > 0) {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setMessages(
          history.map((msg, index) => ({
            id: msg.id || `${sessionId}-${index}`,
            role: msg.role,
            content: msg.content,
            timestamp: new Date(msg.created_at),
            retrievedChunks: msg.retrieved_chunks,
            usage: msg.usage,
          })),
        );
        const totalTokens = history.reduce((acc, msg) => acc + (msg.usage?.total_tokens || 0), 0);
        setSessionTotalTokens(totalTokens);
        setTimeout(scrollToBottom, 50);
      } else {
        setMessages([WELCOME_MESSAGE]);
        setSessionTotalTokens(undefined);
      }
    } else {
      setMessages([WELCOME_MESSAGE]);
      setSessionTotalTokens(undefined);
    }
  }, [sessionId, scrollToBottom]);

  const handleSubmit = async () => {
    const content = inputValue.trim();
    if (!content || isLoading) return;

    if (!folderId) {
      toast.error("请先选择一个文件夹");
      return;
    }

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: "user",
      content,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInputValue("");
    setIsLoading(true);
    scrollToBottom();

    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }

    // 创建新会话 ID
    const newSessionId = currentSessionIdRef.current || generateId();
    if (!currentSessionIdRef.current) {
      currentSessionIdRef.current = newSessionId;
      onSessionCreated(newSessionId);
      mockRagSessions[newSessionId] = [];
    }

    // 保存用户消息
    mockRagSessions[newSessionId].push({
      id: userMessage.id,
      role: "user",
      content: userMessage.content,
      created_at: userMessage.timestamp.toISOString(),
    });

    // AI 消息初始状态
    const assistantMessageId = (Date.now() + 1).toString();
    const assistantMessage: ChatMessage = {
      id: assistantMessageId,
      role: "assistant",
      content: "",
      timestamp: new Date(),
    };

    // 添加空的 AI 消息
    setMessages((prev) => [...prev, assistantMessage]);
    currentAssistantMessageRef.current = assistantMessageId;

    try {
      // 调用真实的 SSE API
      await chatStream(
        {
          query: content,
          folder_id: folderId,
        },
        (event) => {
          // 处理流式事件
          if (event.type === "stream_chunk" && event.content) {
            // 流式内容更新
            setMessages((prev) =>
              prev.map((msg) => {
                if (msg.id === assistantMessageId) {
                  return { ...msg, content: msg.content + event.content! };
                }
                return msg;
              }),
            );
            scrollToBottom();
          } else if (event.type === "tool_result_chunk" && event.content) {
            // 工具调用结果（如 RAG 检索到的 chunk）
            setMessages((prev) =>
              prev.map((msg) => {
                if (msg.id === assistantMessageId) {
                  return { ...msg, content: msg.content + `\n[工具结果] ${event.content}` };
                }
                return msg;
              }),
            );
            scrollToBottom();
          } else if (event.type === "error" && event.content) {
            // 错误消息
            toast.error(event.content);
          }
        },
        () => {
          // 完成
          setIsLoading(false);
          // 保存 AI 回复到会话历史
          setMessages((prev) => {
            const aiMsg = prev.find((m) => m.id === assistantMessageId);
            if (aiMsg) {
              mockRagSessions[newSessionId].push({
                id: aiMsg.id,
                role: "assistant",
                content: aiMsg.content,
                created_at: aiMsg.timestamp.toISOString(),
              });
            }
            return prev;
          });
        },
        (error) => {
          // 错误处理
          setIsLoading(false);
          toast.error(`请求失败: ${error.message}`);
          // 移除失败的 AI 消息
          setMessages((prev) => prev.filter((m) => m.id !== assistantMessageId));
          mockRagSessions[newSessionId] = mockRagSessions[newSessionId].filter(
            (m) => m.id !== userMessage.id,
          );
        },
      );
    } catch {
      setMessages((prev) => prev.filter((m) => m.id !== userMessage.id));
      setIsLoading(false);
      toast.error("发送消息失败");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleTextareaInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(e.target.value);
    // Auto-resize
    const el = e.target;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
  };

  return (
    <div className="h-full flex flex-col bg-background">
      {/* Messages area */}
      <div ref={scrollRef} className="flex-1 overflow-auto p-6 space-y-6">
        {messages.map((message) => (
          <div
            key={message.id}
            className={cn("flex", message.role === "user" ? "justify-end" : "justify-start")}
          >
            <div
              className={cn(
                "max-w-[80%] rounded-lg px-4 py-3 text-sm",
                message.role === "user" ? "bg-primary text-primary-foreground" : "bg-muted",
              )}
            >
              {/* Message content */}
              <div className="whitespace-pre-wrap wrap-break-word">{message.content}</div>

              {/* Retrieved chunks */}
              {message.retrievedChunks && message.retrievedChunks.length > 0 && (
                <div className="mt-3 pt-3 border-t border-border/50">
                  <div className="text-xs text-muted-foreground mb-2">参考来源：</div>
                  <div className="space-y-2">
                    {message.retrievedChunks.map((chunk, index) => (
                      <div
                        key={`${chunk.document_id}-${chunk.chunk_id}`}
                        className="text-xs p-2 rounded bg-background/50"
                      >
                        <div className="font-medium">
                          {index + 1}.{chunk.filename}
                        </div>
                        <div className="text-muted-foreground mt-1 line-clamp-2">
                          {chunk.content}
                        </div>
                        <div className="text-muted-foreground mt-1 text-[10px]">
                          Chunk ID: {chunk.chunk_id} | Document: {chunk.document_id.slice(0, 8)}
                          ...
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Token usage */}
              {message.usage && (
                <div className="mt-2 pt-2 border-t border-border/50">
                  <div className="text-xs text-muted-foreground flex items-center gap-4">
                    <span>Token 用量:</span>
                    <span>
                      输入
                      {message.usage.prompt_tokens}
                    </span>
                    <span>
                      输出
                      {message.usage.completion_tokens}
                    </span>
                    <span>
                      总计
                      {message.usage.total_tokens}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}

        {/* Loading indicator */}
        {isLoading && (
          <div className="flex justify-start">
            <div className="max-w-[80%] rounded-lg px-4 py-3 bg-muted">
              <div className="flex items-center gap-2 text-muted-foreground text-sm">
                <IconLoader2 className="h-4 w-4 animate-spin" />
                <span>正在思考...</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Token consumption bar */}
      {sessionTotalTokens !== undefined && (
        <div className="border-t px-6 py-2 bg-muted/30">
          <div className="max-w-4xl mx-auto flex items-center justify-between text-xs">
            <div className="flex items-center gap-4 text-muted-foreground">
              <span>
                会话 Token 消耗:{" "}
                <strong className="text-foreground">{formatTokenCount(sessionTotalTokens)}</strong>
              </span>
              <span>
                限制:
                {formatTokenCount(CONTEXT_LIMIT)}
              </span>
            </div>
            {sessionTotalTokens >= CONTEXT_LIMIT && (
              <div className="flex items-center gap-2 text-destructive font-medium">
                <span>Token 已超出上下文限制，建议开启新一轮对话</span>
              </div>
            )}
            {sessionTotalTokens >= CONTEXT_WARNING_THRESHOLD &&
              sessionTotalTokens < CONTEXT_LIMIT && (
                <div className="flex items-center gap-2 text-orange-500 font-medium">
                  <span>Token 消耗较高，建议开启新一轮对话</span>
                </div>
              )}
          </div>
        </div>
      )}

      {/* Input area */}
      <div className="border-t p-6">
        <div className="max-w-4xl mx-auto">
          <div className="flex items-end gap-2 rounded-lg border bg-background p-2">
            <textarea
              ref={textareaRef}
              value={inputValue}
              onChange={handleTextareaInput}
              onKeyDown={handleKeyDown}
              placeholder="输入消息... (Enter 发送，Shift+Enter 换行)"
              className="flex-1 resize-none bg-transparent px-2 py-1.5 text-sm outline-none placeholder:text-muted-foreground min-h-9 max-h-50"
              rows={1}
              disabled={isLoading}
            />
            <Button size="icon" onClick={handleSubmit} disabled={isLoading || !inputValue.trim()}>
              {isLoading ? (
                <IconLoader2 className="h-4 w-4 animate-spin" />
              ) : (
                <IconSend className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
