import { useCallback, useRef, useState } from "react";

import { sseRequest } from "@/utils/request/sse";

export interface AgentSession {
  id: string;
  title?: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

export interface AgentToolCall {
  id: string;
  name: string;
  input: Record<string, unknown>;
  output?: unknown;
  error?: string;
}

export interface AgentMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  tool_calls?: AgentToolCall[];
  created_at: string;
}

const TOOL_NAMES: Record<string, string> = {
  create_department: "创建部门",
  update_department: "更新部门",
  delete_department: "删除部门",
  get_department: "查询部门",
  list_departments: "部门列表",
  search_departments: "搜索部门",
  create_role: "创建角色",
  update_role: "更新角色",
  delete_role: "删除角色",
  get_role: "查询角色",
  list_roles: "角色列表",
  create_user: "创建用户",
  update_user: "更新用户",
  delete_user: "删除用户",
  get_user: "查询用户",
  list_users: "用户列表",
};

export function getToolDisplayName(toolName?: string): string {
  return toolName ? TOOL_NAMES[toolName] || toolName : "";
}

function generateId(): string {
  return `agent-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export interface UseAgentChat {
  messages: AgentMessage[];
  loading: boolean;
  sendMessage: (text: string) => Promise<void>;
}

export function useAgentChat(): UseAgentChat {
  const [messages, setMessages] = useState<AgentMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const sendMessage = useCallback(
    async (text: string) => {
      if (!text.trim() || loading) return;

      // 停止之前的请求
      abortControllerRef.current?.abort();
      abortControllerRef.current = new AbortController();

      const userMsg: AgentMessage = {
        id: generateId(),
        role: "user",
        content: text,
        created_at: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, userMsg]);
      setLoading(true);

      let assistantMsg: AgentMessage | null = null;
      let currentToolCalls: AgentToolCall[] = [];

      sseRequest({
        url: "/api/ai/chat",
        method: "POST",
        body: {
          query: text,
        },
        signal: abortControllerRef.current.signal,
        onMessage: (data) => {
          try {
            const parsed = JSON.parse(data);
            switch (parsed.type) {
              case "stream_chunk": {
                // 流式内容
                if (!assistantMsg) {
                  assistantMsg = {
                    id: generateId(),
                    role: "assistant",
                    content: "",
                    created_at: new Date().toISOString(),
                  };
                  setMessages((prev) => [...prev, assistantMsg!]);
                }
                assistantMsg.content += parsed.content;
                setMessages((prev) => {
                  const updated = [...prev];
                  const lastMsg = updated[updated.length - 1];
                  if (lastMsg?.id === assistantMsg?.id) {
                    lastMsg.content = assistantMsg!.content;
                  }
                  return updated;
                });
                break;
              }

              case "tool_result": {
                // 工具调用结果
                if (parsed.tool_calls && Array.isArray(parsed.tool_calls)) {
                  currentToolCalls = parsed.tool_calls.map(
                    (tc: {
                      id?: string;
                      function?: { name?: string; arguments?: string };
                      output?: string;
                    }) => {
                      let input: Record<string, unknown> = {};
                      try {
                        if (tc.function?.arguments) {
                          input = JSON.parse(tc.function.arguments);
                        }
                      } catch {
                        // ignore parse error
                      }
                      let output: unknown = undefined;
                      if (tc.output) {
                        try {
                          output = JSON.parse(tc.output);
                        } catch {
                          output = tc.output;
                        }
                      }
                      return {
                        id: tc.id || generateId(),
                        name: tc.function?.name || "unknown",
                        input,
                        output,
                      };
                    },
                  );
                }
                break;
              }

              case "message": {
                // 最终消息，可能包含工具调用
                if (!assistantMsg) {
                  assistantMsg = {
                    id: generateId(),
                    role: "assistant",
                    content: "",
                    created_at: new Date().toISOString(),
                  };
                  setMessages((prev) => [...prev, assistantMsg!]);
                }
                if (parsed.content) {
                  assistantMsg.content += parsed.content;
                }
                if (parsed.tool_calls && Array.isArray(parsed.tool_calls)) {
                  currentToolCalls = parsed.tool_calls.map(
                    (tc: {
                      id?: string;
                      function?: { name?: string; arguments?: string };
                      output?: string;
                    }) => {
                      let input: Record<string, unknown> = {};
                      try {
                        if (tc.function?.arguments) {
                          input = JSON.parse(tc.function.arguments);
                        }
                      } catch {
                        // ignore parse error
                      }
                      let output: unknown = undefined;
                      if (tc.output) {
                        try {
                          output = JSON.parse(tc.output);
                        } catch {
                          output = tc.output;
                        }
                      }
                      return {
                        id: tc.id || generateId(),
                        name: tc.function?.name || "unknown",
                        input,
                        output,
                      };
                    },
                  );
                }
                break;
              }
            }

            // 更新消息的工具调用
            if (currentToolCalls.length > 0 && assistantMsg) {
              setMessages((prev) => {
                const updated = [...prev];
                const lastMsg = updated[updated.length - 1];
                if (lastMsg?.id === assistantMsg?.id) {
                  lastMsg.tool_calls = [...currentToolCalls];
                }
                return updated;
              });
            }
          } catch {
            // 忽略解析错误
          }
        },
        onError: (error) => {
          const errorMsg: AgentMessage = {
            id: generateId(),
            role: "assistant",
            content: `请求失败：${error.message || "未知错误"}`,
            created_at: new Date().toISOString(),
          };
          setMessages((prev) => [...prev, errorMsg]);
        },
        onComplete: () => {
          setLoading(false);
        },
      });
    },
    [loading],
  );

  return {
    messages,
    loading,
    sendMessage,
  };
}
