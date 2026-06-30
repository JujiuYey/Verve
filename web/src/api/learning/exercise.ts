import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningExercise {
  id: string;
  session_id: string;
  objective_id: string;
  user_id: string;
  type: string;
  prompt: string;
  user_answer?: string;
  verdict?: string;
  mastery_after?: string;
  feedback?: string;
  created_at: string;
  updated_at: string;
}

export interface ExercisePageResponse {
  data: LearningExercise[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const api = {
  page: (page = 1, pageSize = 20) =>
    request.get<ExercisePageResponse>(`${BASE}/exercise/page`, {
      params: { page, page_size: pageSize },
    }),
};

export const exerciseKeys = {
  all: ["learning-exercises"] as const,
  lists: () => [...exerciseKeys.all, "list"] as const,
  list: (page: number, pageSize: number) => [...exerciseKeys.lists(), page, pageSize] as const,
};

export function useExerciseList(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: exerciseKeys.list(page, pageSize),
    queryFn: () => api.page(page, pageSize),
  });
}

export const exerciseApi = api;
