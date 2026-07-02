import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningObjective {
  id: string;
  user_id: string;
  stage_title?: string;
  title: string;
  detail?: string;
  source_document_id?: string;
  source_folder_id?: string;
  source_folder_path?: string;
  order_index: number;
  status: string;
  mastery_level: string;
}

const api = {
  list: (params?: { document_id?: string; folder_id?: string }) =>
    request.get<LearningObjective[]>(`${BASE}/objective/`, { params }),

  detail: (id: string) => request.get<LearningObjective>(`${BASE}/objective/${id}`),
};

export const objectiveKeys = {
  all: ["learning-objectives"] as const,
  list: (params?: { document_id?: string; folder_id?: string }) =>
    [...objectiveKeys.all, "list", params?.document_id || "", params?.folder_id || ""] as const,
  detail: (id: string) => [...objectiveKeys.all, id] as const,
};

export function useObjectives(params?: { document_id?: string; folder_id?: string }) {
  return useQuery({
    queryKey: objectiveKeys.list(params),
    queryFn: () => api.list(params),
    enabled: !!params?.document_id || !!params?.folder_id,
  });
}

export function useObjectiveDetail(id: string) {
  return useQuery({
    queryKey: objectiveKeys.detail(id),
    queryFn: () => api.detail(id),
    enabled: !!id,
  });
}

export const objectiveApi = api;
