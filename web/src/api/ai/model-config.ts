import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

// 模型类型
export type ModelType = "chat" | "embedding" | "rerank";

// 模型配置接口定义
export interface ModelConfig {
  id: string;
  vendor: string;
  name: string;
  api_key: string;
  base_url: string;
  model_type: ModelType;
  model: string;
  temperature: number;
  top_p: number;
  max_tokens?: number;
  top_k?: number;
  is_active: boolean;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateModelConfigRequest {
  vendor: string;
  name: string;
  api_key: string;
  base_url: string;
  model_type: ModelType;
  model: string;
  temperature: number;
  top_p: number;
  max_tokens?: number;
  top_k?: number;
  is_active: boolean;
  is_default: boolean;
}

export interface UpdateModelConfigRequest {
  id: string;
  vendor: string;
  name: string;
  api_key: string;
  base_url: string;
  model_type: ModelType;
  model: string;
  temperature: number;
  top_p: number;
  max_tokens?: number;
  top_k?: number;
  is_active: boolean;
  is_default: boolean;
}

const RESOURCE_PATH = "/api/ai/model-config";

// API 函数
const api = {
  findOne: (id: string) => request.get<ModelConfig>(`${RESOURCE_PATH}/${id}`),

  list: () => request.get<ModelConfig[]>(`${RESOURCE_PATH}/list`),

  create: (data: CreateModelConfigRequest) => request.post<ModelConfig>(RESOURCE_PATH, data),

  update: (data: UpdateModelConfigRequest) => request.put<ModelConfig>(RESOURCE_PATH, data),

  delete: (id: string) => request.delete(`${RESOURCE_PATH}/${id}`),

  setDefault: (id: string) => request.post(`${RESOURCE_PATH}/${id}/default`),
};

// Query keys
export const modelConfigKeys = {
  all: ["model-configs"] as const,
  lists: () => [...modelConfigKeys.all, "list"] as const,
  list: () => [...modelConfigKeys.lists()],
  details: () => [...modelConfigKeys.all, "detail"] as const,
  detail: (id: string) => [...modelConfigKeys.details(), id] as const,
};

// 获取模型配置列表 Hook
export function useModelConfigList() {
  return useQuery({
    queryKey: modelConfigKeys.list(),
    queryFn: () => api.list().then((res) => res || []),
  });
}

// 获取单个模型配置 Hook
export function useModelConfig(id: string) {
  return useQuery({
    queryKey: modelConfigKeys.detail(id),
    queryFn: () => api.findOne(id),
    enabled: !!id,
  });
}

// 创建模型配置 Hook
export function useCreateModelConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateModelConfigRequest) => api.create(data),
    onSuccess: (result) => {
      toast.success("创建成功");
      // 更新列表缓存
      queryClient.setQueryData<ModelConfig[]>(modelConfigKeys.list(), (old) => {
        return old ? [...old, result] : [result];
      });
    },
    onError: () => {
      toast.error("创建失败");
    },
  });
}

// 更新模型配置 Hook
export function useUpdateModelConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateModelConfigRequest) => api.update(data),
    onSuccess: (result) => {
      toast.success("保存成功");
      // 更新列表缓存
      queryClient.setQueryData<ModelConfig[]>(modelConfigKeys.list(), (old) => {
        return old ? old.map((config) => (config.id === result.id ? result : config)) : [result];
      });
    },
    onError: () => {
      toast.error("保存失败");
    },
  });
}

// 删除模型配置 Hook
export function useDeleteModelConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(id),
    onMutate: async (deletedId) => {
      // 乐观更新：立即从列表中移除
      await queryClient.cancelQueries({ queryKey: modelConfigKeys.list() });

      const previousData = queryClient.getQueryData<ModelConfig[]>(modelConfigKeys.list());

      queryClient.setQueryData<ModelConfig[]>(modelConfigKeys.list(), (old) => {
        return old ? old.filter((config) => config.id !== deletedId) : [];
      });

      return { previousData };
    },
    onError: (_err, _id, context) => {
      // 错误时回滚
      if (context?.previousData) {
        queryClient.setQueryData(modelConfigKeys.list(), context.previousData);
      }
      toast.error("删除失败");
    },
    onSuccess: () => {
      toast.success("删除成功");
    },
  });
}

// 设置默认模型配置 Hook
export function useSetDefaultModelConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.setDefault(id),
    onSuccess: () => {
      toast.success("设置默认模型成功");
      // 刷新列表缓存，因为 is_default 在多条记录上都变了
      queryClient.invalidateQueries({ queryKey: modelConfigKeys.list() });
    },
    onError: () => {
      toast.error("设置默认模型失败");
    },
  });
}

// 刷新列表的辅助函数
export function useInvalidateModelConfigList() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: modelConfigKeys.list() });
}

// 导出 API 对象供其他场景使用
export const modelConfigApi = api;
