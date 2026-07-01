import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

import { systemModelKeys } from "./models";

const RESOURCE_PATH = "/api/system/platforms";

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

export interface SyncModelsResult {
  synced: number;
  skipped: number;
  errors: string[];
}

const modelPlatformKeys = {
  all: ["system-model-platforms"] as const,
};

const platformsApi = {
  list: () => request.get<AIPlatform[]>(RESOURCE_PATH),
  create: (data: CreateAIPlatformRequest) => request.post<AIPlatform>(RESOURCE_PATH, data),
  updateConfig: (platformId: string, data: UpdateAIPlatformConfigRequest) =>
    request.put<AIPlatform>(`${RESOURCE_PATH}/${platformId}/config`, data),
  delete: (platformId: string) => request.delete(`${RESOURCE_PATH}/${platformId}`),
  syncModels: (platformId: string) =>
    request.post<SyncModelsResult>(`${RESOURCE_PATH}/${platformId}/sync-models`),
};

export function useModelPlatforms() {
  return useQuery({
    queryKey: modelPlatformKeys.all,
    queryFn: () => platformsApi.list(),
  });
}

export function useCreateAIPlatform() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAIPlatformRequest) => platformsApi.create(data),
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
    mutationFn: (platformId: string) => platformsApi.delete(platformId),
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
    }) => platformsApi.updateConfig(platformId, data),
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
    mutationFn: (platformId: string) => platformsApi.syncModels(platformId),
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