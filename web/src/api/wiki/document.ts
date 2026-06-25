import { request } from "@/utils/request";

// 文档接口类型定义
export interface Document {
  id: string;
  filename: string;
  file_size: number;
  content_type: string;
  file_path: string;
  status: "pending" | "processing" | "completed" | "failed";
  chunk_count: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface PageDocumentsParams {
  page_size?: number;
  page?: number;
  name?: string;
  folder_id?: string;
}

export interface ListDocumentsParams {
  name?: string;
  folder_id?: string;
}

export interface PageDocumentsResponse {
  data: Document[];
  total: number;
  page_size: number;
  page: number;
  total_page: number;
}

export interface DocumentUploadResponse {
  message: string;
  filename: string;
  document_id: string;
  file_path: string;
}

export interface DocumentDownloadResponse {
  download_url: string;
  filename: string;
  expires_in: string;
}

export interface DocumentProcessResponse {
  message: string;
  document_id: string;
  filename: string;
  status: string;
  chunk_count?: number;
}

export interface DocumentContentResponse {
  content: string;
  filename: string;
}

export interface UpdateContentPayload {
  content: string;
}

export interface Chunk {
  ChunkId: string;
  ChunkIndex: number;
  ChunkText: string;
  ChunkSize: number;
  DocumentID: string;
  Filename: string;
  FolderID: string;
  VectorDim: number;
}

export interface ChunksResponse {
  document_id: string;
  chunk_count: number;
  chunks: Chunk[];
}

const RESOURCE_PATH = "/api/wiki/documents";

// 文档相关 API
export const documentApi = {
  // 上传文档
  upload: (file: File, folderId: string) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("folder_id", folderId);

    return request.post<DocumentUploadResponse>(`${RESOURCE_PATH}/upload`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  // 获取文档列表
  page: (params?: PageDocumentsParams) =>
    request.get<PageDocumentsResponse>(`${RESOURCE_PATH}/page`, { params }),

  // 获取文档列表（不分页）
  list: (params?: ListDocumentsParams) =>
    request.get<Document[]>(`${RESOURCE_PATH}/list`, { params }),

  // 获取文档详情
  findOne: (id: string) => request.get<Document>(`${RESOURCE_PATH}/${id}`),

  // 下载文档
  download: (id: string) =>
    request.get<DocumentDownloadResponse>(`${RESOURCE_PATH}/${id}/download`),

  // 删除文档
  delete: (id: string) => request.delete<{ message: string }>(`${RESOURCE_PATH}/${id}`),

  // 处理文档（向量化）
  process: (id: string) =>
    request.post<DocumentProcessResponse>(`${RESOURCE_PATH}/${id}/reprocess`),

  // 获取文档内容
  getContent: (id: string) =>
    request.get<DocumentContentResponse>(`${RESOURCE_PATH}/${id}/content`),

  // 更新文档内容
  updateContent: (id: string, data: UpdateContentPayload) =>
    request.put<{ message: string }>(`${RESOURCE_PATH}/${id}/content`, data),

  // 获取文档 chunks
  getChunks: (id: string) => request.get<ChunksResponse>(`${RESOURCE_PATH}/${id}/chunks`),

  // 删除文档 chunks
  deleteChunks: (id: string) =>
    request.delete<{ message: string }>(`${RESOURCE_PATH}/${id}/chunks`),
};
