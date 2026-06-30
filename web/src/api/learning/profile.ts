import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningProfile {
  id: string;
  user_id: string;
  goal_id: string;
  current_level?: string;
  completed_topics?: string[];
  weak_points?: string[];
  verification_habits?: string;
  next_goal?: string;
  created_at: string;
  updated_at: string;
}

const api = {
  get: (goalId: string) => request.get<LearningProfile | null>(`${BASE}/goal/${goalId}/profile`),
};

export const profileKeys = {
  all: ["learning-profiles"] as const,
  byGoal: (goalId: string) => [...profileKeys.all, goalId] as const,
};

export function useLearningProfile(goalId: string | undefined) {
  return useQuery({
    queryKey: profileKeys.byGoal(goalId ?? ""),
    queryFn: () => api.get(goalId as string),
    enabled: !!goalId,
    retry: false,
  });
}

export const profileApi = api;
