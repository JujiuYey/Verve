const BASE = "/api/learning/ask";
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export interface KnowledgeQAHistoryMessage {
  role: "user" | "assistant";
  content: string;
}

export interface KnowledgeQASource {
  documentId: string;
  documentTitle: string;
  folderPath: string;
  headingPath: string;
  score: number;
}

export type KnowledgeQAStreamEvent =
  | { type: "status"; phase: "generating" }
  | { type: "sources"; sources: KnowledgeQASource[] }
  | { type: "answer"; knowledgeAnswer: string; learningAdvice: string }
  | { type: "error"; message: string };

export interface KnowledgeQAStreamRequest {
  message: string;
  history: KnowledgeQAHistoryMessage[];
  signal?: AbortSignal;
  onEvent: (event: KnowledgeQAStreamEvent) => void;
}

export async function knowledgeQAStream({
  message,
  history,
  signal,
  onEvent,
}: KnowledgeQAStreamRequest): Promise<void> {
  const response = await fetch(`${API_BASE_URL}${BASE}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message, history }),
    signal,
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => null);
    const errorMessage =
      payload && typeof payload === "object" && "message" in payload
        ? String(payload.message)
        : `请求失败: ${response.status}`;
    throw new Error(errorMessage);
  }
  if (!response.body) {
    throw new Error("无法读取回答流");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let completed = false;
  let terminalEventReceived = false;

  try {
    while (!completed) {
      const { done, value } = await reader.read();
      buffer += decoder.decode(value, { stream: !done });
      const lines = buffer.split(/\r?\n/);
      buffer = done ? "" : (lines.pop() ?? "");

      for (const line of lines) {
        if (!line.startsWith("data: ")) continue;
        const data = line.slice(6).trim();
        if (data === "[DONE]") {
          completed = true;
          break;
        }
        if (data) {
          const event = parseKnowledgeQAEvent(data);
          if (event.type === "answer" || event.type === "error") terminalEventReceived = true;
          onEvent(event);
        }
      }

      if (done) break;
    }
  } catch (error) {
    await reader.cancel().catch(() => undefined);
    throw error;
  } finally {
    reader.releaseLock();
  }

  if (!completed) {
    throw new Error("回答连接提前结束，请重试");
  }
  if (!terminalEventReceived) {
    throw new Error("回答未完成，请重试");
  }
}

export function parseKnowledgeQAEvent(data: string): KnowledgeQAStreamEvent {
  const payload: unknown = JSON.parse(data);
  if (!payload || typeof payload !== "object" || !("type" in payload)) {
    throw new Error("知识问答事件格式错误");
  }

  const event = payload as Record<string, unknown>;
  switch (event.type) {
    case "status":
      if (event.phase === "generating") return { type: "status", phase: "generating" };
      break;
    case "sources":
      if (Array.isArray(event.sources) && event.sources.every(isKnowledgeQASource)) {
        return { type: "sources", sources: event.sources };
      }
      break;
    case "answer":
      if (typeof event.knowledgeAnswer === "string" && typeof event.learningAdvice === "string") {
        return {
          type: "answer",
          knowledgeAnswer: event.knowledgeAnswer,
          learningAdvice: event.learningAdvice,
        };
      }
      break;
    case "error":
      if (typeof event.message === "string") return { type: "error", message: event.message };
      break;
  }
  throw new Error("知识问答事件格式错误");
}

function isKnowledgeQASource(value: unknown): value is KnowledgeQASource {
  if (!value || typeof value !== "object") return false;
  const source = value as Record<string, unknown>;
  return (
    typeof source.documentId === "string" &&
    typeof source.documentTitle === "string" &&
    typeof source.folderPath === "string" &&
    typeof source.headingPath === "string" &&
    typeof source.score === "number" &&
    Number.isFinite(source.score)
  );
}
