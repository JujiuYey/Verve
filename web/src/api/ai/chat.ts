import { useAuthStore } from "@/stores/auth";

// API base URL
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// 对话请求参数
export interface ChatRequest {
  query: string;
  folder_id?: string;
  document_id?: string;
}

// 流式事件类型
export interface StreamChunkEvent {
  type: "stream_chunk" | "tool_result_chunk" | "message" | "tool_result" | "action" | "error";
  content?: string;
  agent?: string;
  tool_calls?: unknown[];
  action?: string;
  error?: string;
}

// 回调类型
export type StreamCallback = (event: StreamChunkEvent) => void;

// SSE 流式对话
export async function chatStream(
  params: ChatRequest,
  onMessage: StreamCallback,
  onComplete?: () => void,
  onError?: (error: Error) => void,
): Promise<void> {
  const { query, folder_id, document_id } = params;

  try {
    // 获取 token
    const { accessToken } = useAuthStore.getState();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (accessToken) {
      headers.Authorization = `Bearer ${accessToken}`;
    }

    const response = await fetch(`${API_BASE_URL}/api/ai/chat`, {
      method: "POST",
      headers,
      body: JSON.stringify({
        query,
        folder_id: folder_id || "",
        document_id: document_id || "",
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error("Response body is not readable");
    }

    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();

      if (done) {
        onComplete?.();
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      for (const line of lines) {
        if (line.startsWith("data: ")) {
          const data = line.slice(6).trim();
          if (data === "[DONE]") {
            onComplete?.();
            return;
          }
          try {
            const event = JSON.parse(data) as StreamChunkEvent;
            onMessage(event);
          } catch {
            console.warn("Failed to parse SSE data:", data);
          }
        }
      }
    }
  } catch (error) {
    onError?.(error instanceof Error ? error : new Error(String(error)));
  }
}
