import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

// Collection 接口定义
export interface Collection {
  name: string;
  vectors_count: number;
  points_count: number;
  segments_count: number;
  vector_size: number;
  status: string;
  distance_function: string;
  created_at?: string;
}

// Point 接口定义
export interface Point {
  id: string;
  score?: number;
  payload: Record<string, unknown>;
}

// 创建 Collection 请求
export interface CreateCollectionRequest {
  name: string;
  vector_size: number;
  distance: string;
}

const RESOURCE_PATH = "/api/ai/collections";

// API 函数
const api = {
  list: () => request.get<Collection[]>(RESOURCE_PATH),

  get: (name: string) => request.get<Collection>(`${RESOURCE_PATH}/${name}`),

  create: (data: CreateCollectionRequest) => request.post(RESOURCE_PATH, data),

  delete: (name: string) => request.delete(`${RESOURCE_PATH}/${name}`),

  getPoints: (name: string, limit?: number) =>
    request.get<{ points: Point[] }>(`${RESOURCE_PATH}/${name}/points?limit=${limit || 20}`),

  getStats: (name: string) =>
    request.get<Record<string, unknown>>(`${RESOURCE_PATH}/${name}/stats`),
};

// Query keys
export const collectionKeys = {
  all: ["collections"] as const,
  lists: () => [...collectionKeys.all, "list"] as const,
  list: () => [...collectionKeys.lists()],
  details: () => [...collectionKeys.all, "detail"] as const,
  detail: (name: string) => [...collectionKeys.details(), name] as const,
};

// 获取 Collection 列表 Hook
export function useCollectionList() {
  return useQuery({
    queryKey: collectionKeys.list(),
    queryFn: () => api.list().then((res) => res || []),
  });
}

// 获取单个 Collection 详情 Hook
export function useCollection(name: string) {
  return useQuery({
    queryKey: collectionKeys.detail(name),
    queryFn: () => api.get(name),
    enabled: !!name,
  });
}

// 创建 Collection Hook
export function useCreateCollection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateCollectionRequest) => api.create(data),
    onSuccess: () => {
      toast.success("创建成功");
      queryClient.invalidateQueries({ queryKey: collectionKeys.list() });
    },
    onError: () => {
      toast.error("创建失败");
    },
  });
}

// 删除 Collection Hook
export function useDeleteCollection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => api.delete(name),
    onSuccess: () => {
      toast.success("删除成功");
      queryClient.invalidateQueries({ queryKey: collectionKeys.list() });
    },
    onError: () => {
      toast.error("删除失败");
    },
  });
}

// 获取 Points Hook
export function useCollectionPoints(name: string, limit?: number) {
  return useQuery({
    queryKey: [...collectionKeys.detail(name), "points", limit] as const,
    queryFn: () => api.getPoints(name, limit),
    enabled: !!name,
  });
}

// 刷新列表的辅助函数
export function useInvalidateCollectionList() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: collectionKeys.list() });
}

// 导出 API 对象
export const collectionApi = api;
