import { useMutation, useQuery } from "@tanstack/react-query";

import { request } from "@/utils/request";

export type IndexJobStatus = "pending" | "running" | "completed" | "failed" | "superseded";

export interface IndexJobProgress {
  id: string;
  document_id: string;
  document_version: number;
  root_folder_id?: string;
  status: IndexJobStatus;
  error_message?: string;
  chunk_count: number;
  created_at: string;
  started_at?: string;
  finished_at?: string;
}

const RESOURCE_PATH = "/api/rag/wiki";

export const ragWikiApi = {
  listJobs: (rootFolderId?: string) =>
    request.get<IndexJobProgress[]>(`${RESOURCE_PATH}/jobs`, {
      params: rootFolderId ? { root_folder_id: rootFolderId } : undefined,
    }),
  indexDocument: (documentId: string) =>
    request.post<void>(`${RESOURCE_PATH}/documents/${documentId}/index`),
  documentIndexStatus: (documentId: string) =>
    request.get<IndexJobProgress>(`${RESOURCE_PATH}/documents/${documentId}/index-status`),
};

export const ragWikiKeys = {
  documentStatus: (documentId: string) => ["rag-wiki", "document-status", documentId] as const,
};

export function useDocumentIndexStatus(documentId: string) {
  return useQuery({
    queryKey: ragWikiKeys.documentStatus(documentId),
    queryFn: () => ragWikiApi.documentIndexStatus(documentId),
    enabled: !!documentId,
    retry: false,
  });
}

export function useRetryDocumentIndex() {
  return useMutation({ mutationFn: (documentId: string) => ragWikiApi.indexDocument(documentId) });
}
