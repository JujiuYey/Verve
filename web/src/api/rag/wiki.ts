import { request } from "@/utils/request";

export type IndexJobStatus = "pending" | "running" | "completed" | "failed";

export interface IndexJobProgress {
  id: string;
  document_id: string;
  root_folder_id?: string;
  status: IndexJobStatus;
  error_message?: string;
  chunk_count: number;
  created_at: string;
  started_at?: string;
  finished_at?: string;
}

export interface WikiKnowledgeSearchRequest {
  root_folder_id: string;
  query: string;
  limit?: number;
}

export interface WikiKnowledgeSearchResult {
  chunk_id: string;
  score: number;
  root_folder_id: string;
  folder_id: string;
  document_id: string;
  document_title: string;
  folder_path: string;
  heading_path: string;
  content: string;
}

const RESOURCE_PATH = "/api/rag/wiki";

export const ragWikiApi = {
  listJobs: (rootFolderId?: string) =>
    request.get<IndexJobProgress[]>(`${RESOURCE_PATH}/jobs`, {
      params: rootFolderId ? { root_folder_id: rootFolderId } : undefined,
    }),
  indexDocument: (documentId: string) =>
    request.post<void>(`${RESOURCE_PATH}/documents/${documentId}/index`),
  search: (data: WikiKnowledgeSearchRequest) =>
    request.post<WikiKnowledgeSearchResult[]>(`${RESOURCE_PATH}/search`, data),
};
