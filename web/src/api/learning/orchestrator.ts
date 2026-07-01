import { useMutation, useQuery } from "@tanstack/react-query";

import type { LearningExercise } from "@/api/learning/exercise";
import { request } from "@/utils/request";

const BASE = "/api/learning";

export type LearningOrchestratorActionType =
  | "create_goal"
  | "continue_objective"
  | "review_objective"
  | "open_goal";

export interface LearningOrchestratorAction {
  id: string;
  type: LearningOrchestratorActionType;
  priority: number;
  label: string;
  title: string;
  description: string;
  goal_id?: string;
  objective_id?: string;
  intent?: string;
  reason: string;
}

export interface OrchestrateLearningRequest {
  intent?: string;
}

export interface OrchestrateLearningResponse {
  intent: string;
  summary: string;
  habit_summary: string;
  actions: LearningOrchestratorAction[];
  recent: LearningExercise[];
}

const api = {
  orchestrate: (data: OrchestrateLearningRequest = {}) =>
    request.post<OrchestrateLearningResponse>(`${BASE}/orchestrate`, data),
};

export const orchestratorKeys = {
  all: ["learning-orchestrator"] as const,
  home: () => [...orchestratorKeys.all, "home"] as const,
};

export function useLearningOrchestrator() {
  return useQuery({
    queryKey: orchestratorKeys.home(),
    queryFn: () => api.orchestrate(),
  });
}

export function useOrchestrateLearning() {
  return useMutation({
    mutationFn: (data: OrchestrateLearningRequest) => api.orchestrate(data),
  });
}

export const orchestratorApi = api;
