import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningMemoryItem {
  id: string;
  kind: string;
  statement: string;
  confidence: string;
  folder_id?: string | null;
  document_id?: string | null;
  last_seen_at: string;
}

export interface LearningMemoryResponse {
  summary: string;
  highlights: string[];
  items: LearningMemoryItem[];
}

const api = {
  get: (folderId?: string) =>
    request.get<LearningMemoryResponse>(`${BASE}/memory`, {
      params: folderId ? { folder_id: folderId } : undefined,
    }),
};

export function useLearningMemory(folderId?: string) {
  return useQuery({
    queryKey: ["learning-memory", folderId ?? "all"],
    queryFn: () => api.get(folderId),
  });
}

export const memoryApi = api;
