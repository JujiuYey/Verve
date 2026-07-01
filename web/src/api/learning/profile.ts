import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningProfile {
  id: string;
  user_id: string;
  folder_id: string;
  current_level?: string;
  completed_topics?: string[];
  weak_points?: string[];
  verification_habits?: string;
  next_goal?: string;
  created_at: string;
  updated_at: string;
}

const api = {
  get: (folderId: string) =>
    request.get<LearningProfile | null>(`${BASE}/folder/${folderId}/profile`),
};

export const profileKeys = {
  all: ["learning-profiles"] as const,
  byFolder: (folderId: string) => [...profileKeys.all, folderId] as const,
};

export function useLearningProfile(folderId: string | undefined) {
  return useQuery({
    queryKey: profileKeys.byFolder(folderId ?? ""),
    queryFn: () => api.get(folderId as string),
    enabled: !!folderId,
    retry: false,
  });
}

export const profileApi = api;
