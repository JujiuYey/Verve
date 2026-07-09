import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { request } from "@/utils/request";

import type { AIModel } from "./models";

const RESOURCE_PATH = "/api/system/agent-model-configs";

export interface AgentModelConfig {
  id: string;
  agent_key: string;
  scene_key: string;
  model_id: string;
  params?: Record<string, unknown>;
  enabled: boolean;
  model?: AIModel;
  created_at?: string;
  updated_at?: string;
}

export interface UpsertAgentModelConfigRequest {
  model_id: string;
  params?: Record<string, unknown>;
  enabled?: boolean;
}

export const agentModelConfigKeys = {
  all: ["agent-model-configs"] as const,
};

const agentModelConfigsApi = {
  list: () => request.get<AgentModelConfig[]>(RESOURCE_PATH),
  upsert: (agentKey: string, sceneKey: string, data: UpsertAgentModelConfigRequest) =>
    request.put<AgentModelConfig>(`${RESOURCE_PATH}/${agentKey}/${sceneKey}`, data),
};

export function useAgentModelConfigs() {
  return useQuery({
    queryKey: agentModelConfigKeys.all,
    queryFn: () => agentModelConfigsApi.list(),
  });
}

export function useUpsertAgentModelConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      agentKey,
      sceneKey,
      data,
    }: {
      agentKey: string;
      sceneKey: string;
      data: UpsertAgentModelConfigRequest;
    }) => agentModelConfigsApi.upsert(agentKey, sceneKey, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentModelConfigKeys.all });
      toast.success("保存成功");
    },
    onError: () => {
      toast.error("保存失败");
    },
  });
}
