import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

export type ModelType = "chat" | "embedding" | "rerank";
export type ModelCapability = "vision" | "reasoning" | "tool" | "embedding" | "rerank";

export interface AIPlatform {
  id: string;
  name: string;
  provider_type: string;
  default_base_url: string;
  base_url: string;
  api_key_hint?: string;
  api_key_url?: string;
  docs_url?: string;
  model_list_path: string;
  auth_scheme: string;
  enabled: boolean;
  sort_order: number;
  last_model_sync_at?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AIModel {
  id: string;
  platform_id: string;
  model_name: string;
  display_name: string;
  model_type: ModelType;
  capabilities: ModelCapability[];
  status: "active" | "inactive";
  source: "manual" | "remote";
  is_default?: boolean;
  temperature?: number;
  top_p?: number;
  max_tokens?: number;
  top_k?: number;
  created_at?: string;
  updated_at?: string;
}

export interface CreateAIPlatformRequest {
  name: string;
  provider_type: string;
  default_base_url: string;
  base_url: string;
  api_key: string;
  model_list_path: string;
  auth_scheme: string;
  docs_url?: string;
  api_key_url?: string;
}

export interface UpdateAIPlatformConfigRequest {
  base_url: string;
  api_key?: string;
  clear_api_key?: boolean;
}

export interface CreateAIModelRequest {
  platform_id: string;
  model_name: string;
  display_name: string;
  model_type: ModelType;
  capabilities: ModelCapability[];
  source?: "manual" | "remote";
  is_default?: boolean;
  temperature?: number;
  top_p?: number;
  max_tokens?: number;
  top_k?: number;
}

export interface UpdateAIModelRequest {
  status?: "active" | "inactive";
  display_name?: string;
  capabilities?: ModelCapability[];
}

export interface SyncModelsResult {
  synced: number;
  skipped: number;
  errors: string[];
}

const RESOURCE_PATH = "/api/ai/model-config";

const modelPlatformKeys = {
  all: ["system-model-platforms"] as const,
};

const systemModelKeys = {
  all: ["system-models"] as const,
};

const systemModelConfigApi = {
  platforms: () => request.get<AIPlatform[]>(`${RESOURCE_PATH}/platforms`),
  createPlatform: (data: CreateAIPlatformRequest) =>
    request.post<AIPlatform>(`${RESOURCE_PATH}/platforms`, data),
  updatePlatformConfig: (platformId: string, data: UpdateAIPlatformConfigRequest) =>
    request.put<AIPlatform>(`${RESOURCE_PATH}/platforms/${platformId}/config`, data),
  deletePlatform: (platformId: string) =>
    request.delete(`${RESOURCE_PATH}/platforms/${platformId}`),
  syncModels: (platformId: string) =>
    request.post<SyncModelsResult>(`${RESOURCE_PATH}/platforms/${platformId}/sync-models`),
  models: () => request.get<AIModel[]>(`${RESOURCE_PATH}/models`),
  createModel: (data: CreateAIModelRequest) =>
    request.post<AIModel>(`${RESOURCE_PATH}/models`, data),
  updateModel: (modelId: string, data: UpdateAIModelRequest) =>
    request.put<AIModel>(`${RESOURCE_PATH}/models/${modelId}`, data),
  deleteModel: (modelId: string) => request.delete(`${RESOURCE_PATH}/models/${modelId}`),
};

export function useModelPlatforms() {
  return useQuery({
    queryKey: modelPlatformKeys.all,
    queryFn: () => systemModelConfigApi.platforms(),
  });
}

export function useAIModels() {
  return useQuery({
    queryKey: systemModelKeys.all,
    queryFn: () => systemModelConfigApi.models(),
  });
}

export function useCreateAIPlatform() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAIPlatformRequest) => systemModelConfigApi.createPlatform(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelPlatformKeys.all });
      toast.success("创建成功");
    },
    onError: () => {
      toast.error("创建失败");
    },
  });
}

export function useDeleteAIPlatform() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (platformId: string) => systemModelConfigApi.deletePlatform(platformId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelPlatformKeys.all });
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("删除成功");
    },
    onError: () => {
      toast.error("删除失败");
    },
  });
}

export function useUpdateAIPlatformConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      platformId,
      data,
    }: {
      platformId: string;
      data: UpdateAIPlatformConfigRequest;
    }) => systemModelConfigApi.updatePlatformConfig(platformId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelPlatformKeys.all });
      toast.success("保存成功");
    },
    onError: () => {
      toast.error("保存失败");
    },
  });
}

export function useSyncModels() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (platformId: string) => systemModelConfigApi.syncModels(platformId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelPlatformKeys.all });
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("同步成功");
    },
    onError: () => {
      toast.error("同步失败");
    },
  });
}

export function useCreateAIModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAIModelRequest) => systemModelConfigApi.createModel(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("添加成功");
    },
    onError: () => {
      toast.error("添加失败");
    },
  });
}

export function useUpdateAIModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ modelId, data }: { modelId: string; data: UpdateAIModelRequest }) =>
      systemModelConfigApi.updateModel(modelId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("保存成功");
    },
    onError: () => {
      toast.error("保存失败");
    },
  });
}

export function useDeleteAIModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (modelId: string) => systemModelConfigApi.deleteModel(modelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("删除成功");
    },
    onError: () => {
      toast.error("删除失败");
    },
  });
}
