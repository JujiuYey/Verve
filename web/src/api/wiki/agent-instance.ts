import { request } from "@/utils/request";

import type { Folder } from "./folder";

export interface WikiAgentInstance {
  id: string;
  user_id: string;
  root_folder_id: string;
  agent_key: string;
  name: string;
  description?: string;
  status: "active" | "archived";
  created_at: string;
  updated_at: string;
  root_folder?: Folder;
}

export interface EnsureWikiAgentInstanceRequest {
  root_folder_id: string;
  name?: string;
  description?: string;
}

const RESOURCE_PATH = "/api/wiki/agent-instances";

export const wikiAgentInstanceApi = {
  findByRoot: (rootFolderId: string) =>
    request.get<WikiAgentInstance>(RESOURCE_PATH, {
      params: { root_folder_id: rootFolderId },
    }),
  ensure: (data: EnsureWikiAgentInstanceRequest) =>
    request.post<WikiAgentInstance>(`${RESOURCE_PATH}/ensure`, data),
};
