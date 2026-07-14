// ========== SSE 请求封装 ==========

export type SSECallback = (data: string) => void;

export interface SSERequestOptions {
  url: string;
  method?: "POST" | "GET";
  headers?: Record<string, string>;
  body?: unknown;
  onMessage: SSECallback;
  onError?: (error: Error) => void;
  onComplete?: () => void;
  signal?: AbortSignal;
}

/**
 * SSE 请求封装,支持流式读取 Server-Sent Events。
 * 单租户自部署下不再注入 Authorization 头。
 */
export function sseRequest(options: SSERequestOptions): () => void {
  const {
    url,
    method = "POST",
    headers = {},
    body,
    onMessage,
    onError,
    onComplete,
    signal,
  } = options;

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

  const controller = new AbortController();
  const finalSignal = signal ? signal : controller.signal;

  const requestHeaders: Record<string, string> = {
    "Content-Type": "application/json",
    ...headers,
  };

  let buffer = "";
  let reader: ReadableStreamDefaultReader<Uint8Array> | null = null;

  async function startRequest() {
    try {
      const response = await fetch(`${apiBaseUrl}${url}`, {
        method,
        headers: requestHeaders,
        body: body ? JSON.stringify(body) : undefined,
        signal: finalSignal,
      });

      if (!response.ok) {
        throw new Error(`请求失败: ${response.status}`);
      }

      if (!response.body) {
        throw new Error("无法读取响应流");
      }

      reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const data = line.slice(6).trim();
            if (data && data !== "[DONE]") {
              onMessage(data);
            }
          }
        }
      }

      onComplete?.();
    } catch (error) {
      if ((error as Error).name === "AbortError") {
        onComplete?.();
        return;
      }
      onError?.(error as Error);
    }
  }

  startRequest();

  // 返回取消函数
  return () => {
    controller.abort();
    reader?.cancel();
  };
}
