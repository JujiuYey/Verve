import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

// 学习目标
export interface LearningGoal {
  id: string;
  user_id: string;
  title: string;
  description?: string;
  source: string;
  status: string; // active / archived / completed
  created_at: string;
  updated_at: string;
}

// 小目标
export interface LearningObjective {
  id: string;
  path_id: string;
  user_id: string;
  stage_title?: string;
  title: string;
  detail?: string;
  order_index: number;
  status: string; // pending / active / completed / review
  mastery_level: string; // none / seen / heard / explained / written / verified
}

// 学习路线
export interface LearningPath {
  id: string;
  goal_id: string;
  overview?: { title: string }[];
  current_objective_id?: string | null;
  status: string;
}

// 目标详情(含路线 + 进度)
export interface GoalDetail {
  goal: LearningGoal;
  path?: LearningPath;
  objectives?: LearningObjective[];
  current_objective_id?: string | null;
  progress?: { completed: number; total: number };
}

// 继续上次
export interface ContinueInfo {
  goal_id: string;
  objective_id?: string;
  title?: string;
}

export interface CreateGoalRequest {
  title: string;
}

export interface UpdateGoalRequest {
  id: string;
  title?: string;
  status?: string;
}

export interface GoalPageResponse {
  data: LearningGoal[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const BASE = "/api/learning";

const api = {
  // 创建目标(后端会同步生成学习路线)
  create: (data: CreateGoalRequest) => request.post<{ goal_id: string }>(`${BASE}/goal`, data),

  page: (page = 1, pageSize = 20) =>
    request.get<GoalPageResponse>(`${BASE}/goal/page`, { params: { page, page_size: pageSize } }),

  detail: (id: string) => request.get<GoalDetail>(`${BASE}/goal/${id}`),

  update: (data: UpdateGoalRequest) => request.put(`${BASE}/goal`, data),

  delete: (id: string) => request.delete(`${BASE}/goal/${id}`),

  continue: () => request.get<ContinueInfo | null>(`${BASE}/continue`),
};

export const goalKeys = {
  all: ["learning-goals"] as const,
  lists: () => [...goalKeys.all, "list"] as const,
  list: (page: number, pageSize: number) => [...goalKeys.lists(), page, pageSize] as const,
  details: () => [...goalKeys.all, "detail"] as const,
  detail: (id: string) => [...goalKeys.details(), id] as const,
  continue: () => [...goalKeys.all, "continue"] as const,
};

export function useGoalList(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: goalKeys.list(page, pageSize),
    queryFn: () => api.page(page, pageSize),
  });
}

export function useGoalDetail(id: string) {
  return useQuery({
    queryKey: goalKeys.detail(id),
    queryFn: () => api.detail(id),
    enabled: !!id,
  });
}

export function useContinue() {
  return useQuery({
    queryKey: goalKeys.continue(),
    queryFn: () => api.continue(),
  });
}

export function useCreateGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateGoalRequest) => api.create(data),
    onSuccess: () => {
      toast.success("学习目标已创建,路线已生成");
      queryClient.invalidateQueries({ queryKey: goalKeys.lists() });
      queryClient.invalidateQueries({ queryKey: goalKeys.continue() });
    },
    // 失败时 request 层已弹 toast
  });
}

export function useUpdateGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateGoalRequest) => api.update(data),
    onSuccess: () => {
      toast.success("已更新");
      queryClient.invalidateQueries({ queryKey: goalKeys.all });
    },
  });
}

export function useDeleteGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(id),
    onSuccess: () => {
      toast.success("已删除");
      queryClient.invalidateQueries({ queryKey: goalKeys.lists() });
    },
  });
}

export const goalApi = api;
