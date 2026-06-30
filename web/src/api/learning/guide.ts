import { useMutation, useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface GenerateGuideRequest {
  objective_id: string;
  markdown: string;
}

export interface GuideResult {
  summary: string;
  mastery_goals: string[];
  practice_points?: GuidePracticePoint[];
  reading_steps: string[];
  pitfalls: string[];
  self_check_questions: string[];
  evidence: string[];
  content_hash?: string;
  cached?: boolean;
}

export interface GuidePracticePoint {
  title: string;
  goal: string;
  evidence?: string[];
}

const api = {
  get: (objectiveId: string, contentHash: string) =>
    request.get<GuideResult | null>(`${BASE}/guide/${objectiveId}`, {
      params: { content_hash: contentHash },
    }),

  generate: (data: GenerateGuideRequest) =>
    request.post<GuideResult>(`${BASE}/guide/generate`, data),
};

export const guideKeys = {
  all: ["learning-guide"] as const,
  detail: (objectiveId: string, contentHash: string) =>
    [...guideKeys.all, objectiveId, contentHash] as const,
};

export function useGuideCache(objectiveId: string, contentHash: string) {
  return useQuery({
    queryKey: guideKeys.detail(objectiveId, contentHash),
    queryFn: () => api.get(objectiveId, contentHash),
    enabled: !!objectiveId && !!contentHash,
  });
}

export function useGenerateGuide() {
  return useMutation({
    mutationFn: (data: GenerateGuideRequest) => api.generate(data),
  });
}

export const guideApi = api;
