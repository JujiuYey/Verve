import { request } from "@/utils/request";

const RESOURCE_PATH = "/api/rag/wiki";

export interface IndexFolderResponse {
  root_folder_id: string;
  document_count: number;
  started_at: string;
}

export interface IndexJobProgress {
  id: string;
  document_id: string;
  root_folder_id?: string;
  status: "pending" | "running" | "completed" | "failed";
  error_message?: string;
  chunk_count: number;
  created_at: string;
  started_at?: string;
  finished_at?: string;
}

export const ragApi = {
  indexFolder: (folderId: string) =>
    request.post<IndexFolderResponse>(`${RESOURCE_PATH}/folders/${folderId}/index`),

  listJobs: (rootFolderId?: string) =>
    request.get<IndexJobProgress[]>(`${RESOURCE_PATH}/jobs`, {
      params: rootFolderId ? { root_folder_id: rootFolderId } : undefined,
    }),
};
