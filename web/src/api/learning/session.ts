import { useMutation, useQuery } from "@tanstack/react-query";

import { useAuthStore } from "@/stores/auth";
import { request } from "@/utils/request";

const BASE = "/api/learning";
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

export interface LearningMessage {
  id: string;
  session_id: string;
  turn_id: string;
  role: string;
  content: string;
  created_at: string;
}

export interface LearningSession {
  id: string;
  user_id: string;
  document_id: string;
  status: "active" | "completed" | "abandoned";
  summary?: string;
  started_at: string;
  ended_at?: string;
  created_at: string;
  updated_at: string;
}

export interface FeynmanReview {
  heard_summary: string;
  clear_points: string[];
  confusing_points: string[];
  misconceptions: string[];
  follow_up_question: string;
  explanation_summary: string;
  ready_to_wrap_up: boolean;
  context_sufficient: boolean;
}

export interface LearningExplanationReview extends FeynmanReview {
  id: string;
  turn_id: string;
  session_id: string;
  document_id: string;
  user_id: string;
  explanation: string;
  created_at: string;
}

export interface SessionDetail {
  session: LearningSession;
  messages: LearningMessage[];
  reviews: LearningExplanationReview[];
}

export interface CreateSessionRequest {
  document_id: string;
}

export interface ReviewExplanationRequest {
  request_id: string;
  explanation: string;
}

export interface CompleteResult {
  summary: string;
}

export interface LearningStreamEvent {
  type: "stream_chunk" | "message" | "reasoning" | "tool_call" | "tool_result" | "action" | "error";
  content?: string;
  agent?: string;
  phase?: string;
  action?: LearningCoachAction;
  tool_call_id?: string;
  tool_name?: string;
  id?: string;
  name?: string;
  arguments?: string;
}

export interface LearningCoachAction {
  type: "navigate_to_practice";
  document_id?: string;
  label?: string;
}

export interface CoachChatOptions {
  root_folder_id?: string;
}

const api = {
  create: (data: CreateSessionRequest) =>
    request.post<{ session_id: string }>(`${BASE}/session`, data),

  detail: (id: string) => request.get<SessionDetail>(`${BASE}/session/${id}`),

  review: (id: string, data: ReviewExplanationRequest) =>
    request.post<FeynmanReview>(`${BASE}/session/${id}/review`, data),

  complete: (id: string) => request.post<CompleteResult>(`${BASE}/session/${id}/complete`),
};

export const sessionKeys = {
  all: ["learning-session"] as const,
  detail: (id: string) => [...sessionKeys.all, id] as const,
};

export function useSessionDetail(id: string) {
  return useQuery({
    queryKey: sessionKeys.detail(id),
    queryFn: () => api.detail(id),
    enabled: !!id,
  });
}

export function useCreateSession() {
  return useMutation({
    mutationFn: (data: CreateSessionRequest) => api.create(data),
  });
}

export function useReviewExplanation(sessionId: string) {
  return useMutation({
    mutationFn: (data: ReviewExplanationRequest) => api.review(sessionId, data),
  });
}

export function useCompleteSession(sessionId: string) {
  return useMutation({
    mutationFn: () => api.complete(sessionId),
  });
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
        if (!line.startsWith("data: ")) continue;
        const data = line.slice(6).trim();
        if (data === "[DONE]") {
          onComplete?.();
          return;
        }
        try {
          onMessage(JSON.parse(data) as LearningStreamEvent);
        } catch {
          console.warn("Failed to parse SSE data:", data);
        }
      }
    }
  } catch (error) {
    onError?.(error instanceof Error ? error : new Error(String(error)));
  }
}

export const sessionApi = api;
