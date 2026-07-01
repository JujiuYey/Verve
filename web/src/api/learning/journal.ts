import { useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

const BASE = "/api/learning";

export interface LearningJournal {
  id: string;
  user_id: string;
  folder_id: string;
  date: string;
  learned?: string;
  evidence?: string;
  weak_points?: string;
  next_step?: string;
  created_at: string;
}

export interface JournalPageResponse {
  data: LearningJournal[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const api = {
  page: (page = 1, pageSize = 20) =>
    request.get<JournalPageResponse>(`${BASE}/journal/page`, {
      params: { page, page_size: pageSize },
    }),
};

export function useJournalList(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["learning-journals", page, pageSize],
    queryFn: () => api.page(page, pageSize),
  });
}

export const journalApi = api;
