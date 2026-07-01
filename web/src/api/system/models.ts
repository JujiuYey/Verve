import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

const RESOURCE_PATH = "/api/system/models";

export type ModelType = "chat" | "embedding" | "rerank";
export type ModelCapability = "vision" | "reasoning" | "tool" | "embedding" | "rerank";

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

export const systemModelKeys = {
  all: ["system-models"] as const,
};

const modelsApi = {
  list: () => request.get<AIModel[]>(RESOURCE_PATH),
  create: (data: CreateAIModelRequest) => request.post<AIModel>(RESOURCE_PATH, data),
  update: (modelId: string, data: UpdateAIModelRequest) =>
    request.put<AIModel>(`${RESOURCE_PATH}/${modelId}`, data),
  delete: (modelId: string) => request.delete(`${RESOURCE_PATH}/${modelId}`),
};

export function useAIModels() {
  return useQuery({
    queryKey: systemModelKeys.all,
    queryFn: () => modelsApi.list(),
  });
}

export function useCreateAIModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAIModelRequest) => modelsApi.create(data),
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
      modelsApi.update(modelId, data),
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
    mutationFn: (modelId: string) => modelsApi.delete(modelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemModelKeys.all });
      toast.success("删除成功");
    },
    onError: () => {
      toast.error("删除失败");
    },
  });
}