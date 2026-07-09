import { useMutation, useQuery } from "@tanstack/react-query";

import { useAuthStore } from "@/stores/auth";
import { request } from "@/utils/request";

const BASE = "/api/learning";
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

export interface LearningMessage {
  id: string;
  session_id: string;
  role: string; // user / assistant / system
  agent_type?: string;
  content: string;
  created_at: string;
}

export interface LearningSession {
  id: string;
  user_id: string;
  objective_id: string;
  status: string;
  summary?: string;
  started_at: string;
  ended_at?: string;
}

export interface SessionDetail {
  session: LearningSession;
  messages: LearningMessage[];
}

export interface CreateSessionRequest {
  objective_id: string;
}

export interface SubmitExerciseRequest {
  type: string; // explain / choice / cloze / paste_output / code_snippet
  prompt: string;
  user_answer: string;
}

export interface ExerciseResult {
  verdict: string; // pass / partial / fail
  mastery_after: string;
  feedback: string;
  objective_id: string;
  evidence?: string;
  weak_points?: string[];
  next_recommendation?: string;
  review_required?: boolean;
}

export interface CompleteResult {
  summary: string;
  next_objective?: { id: string; title: string };
}

// 陪练 SSE 事件(对齐后端 learning SSE)
export interface LearningStreamEvent {
  type:
    | "stream_chunk"
    | "message"
    | "reasoning"
    | "tool_call"
    | "tool_result"
    | "exercise"
    | "action"
    | "error";
  content?: string;
  agent?: string;
  phase?: string;
  action?: LearningCoachAction;
  // tool_call / tool_result 字段
  tool_call_id?: string;
  tool_name?: string;
  id?: string;
  name?: string;
  arguments?: string;
}

export interface LearningCoachAction {
  type: "navigate_to_practice";
  objective_id?: string;
  folder_id?: string;
  label?: string;
}

export interface CoachChatOptions {
  agent_instance_id?: string;
  root_folder_id?: string;
}

const api = {
  create: (data: CreateSessionRequest) =>
    request.post<{ session_id: string }>(`${BASE}/session`, data),

  detail: (id: string) => request.get<SessionDetail>(`${BASE}/session/${id}`),

  exercise: (id: string, data: SubmitExerciseRequest) =>
    request.post<ExerciseResult>(`${BASE}/session/${id}/exercise`, data),

  complete: (id: string) => request.post<CompleteResult>(`${BASE}/session/${id}/complete`),
};

export function useSessionDetail(id: string) {
  return useQuery({
    queryKey: ["learning-session", id],
    queryFn: () => api.detail(id),
    enabled: !!id,
  });
}

export function useCreateSession() {
  return useMutation({
    mutationFn: (data: CreateSessionRequest) => api.create(data),
  });
}

export function useSubmitExercise(sessionId: string) {
  return useMutation({
    mutationFn: (data: SubmitExerciseRequest) => api.exercise(sessionId, data),
  });
}

export function useCompleteSession(sessionId: string) {
  return useMutation({
    mutationFn: () => api.complete(sessionId),
  });
}

// 陪练对话 SSE(复用 chat.ts 的 fetch + ReadableStream 模式)
export async function sessionChatStream(
  sessionId: string,
  message: string,
  onMessage: (event: LearningStreamEvent) => void,
  onComplete?: () => void,
  onError?: (error: Error) => void,
): Promise<void> {
  try {
    const { accessToken } = useAuthStore.getState();
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (accessToken) {
      headers.Authorization = `Bearer ${accessToken}`;
    }

    const response = await fetch(`${API_BASE_URL}${BASE}/session/${sessionId}/chat`, {
      method: "POST",
      headers,
      body: JSON.stringify({ message }),
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
            const event = JSON.parse(data) as LearningStreamEvent;
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

export async function coachChatStream(
  message: string,
  onMessage: (event: LearningStreamEvent) => void,
  onComplete?: () => void,
  onError?: (error: Error) => void,
  options?: CoachChatOptions,
): Promise<void> {
  try {
    const { accessToken } = useAuthStore.getState();
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (accessToken) {
      headers.Authorization = `Bearer ${accessToken}`;
    }

    const response = await fetch(`${API_BASE_URL}${BASE}/coach/chat`, {
      method: "POST",
      headers,
      body: JSON.stringify({
        message,
        agent_instance_id: options?.agent_instance_id,
        root_folder_id: options?.root_folder_id,
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
            const event = JSON.parse(data) as LearningStreamEvent;
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

export const sessionApi = api;
