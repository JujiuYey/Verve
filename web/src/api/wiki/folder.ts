import { request } from "@/utils/request";

// 文件夹接口定义
export interface Folder {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// 文件夹树节点
export interface FolderTreeNode {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
  hasChildren: boolean;
  children: FolderTreeNode[];
}

export interface CreateFolderRequest {
  name: string;
  description?: string;
  parent_id?: string;
  sort_order?: number;
}

export interface UpdateFolderRequest {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  sort_order?: number;
}

export interface FolderPageResponse {
  data: Folder[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const RESOURCE_PATH = "/api/wiki/folders";

// 文件夹相关 API
export const folderApi = {
  // 获取文件夹详情
  findOne: (id: string) => request.get<Folder>(`${RESOURCE_PATH}/${id}`),

  // 获取文件夹列表
  list: (parentId?: string) =>
    request.get<Folder[]>(`${RESOURCE_PATH}/list`, {
      params: parentId ? { parent_id: parentId } : undefined,
    }),

  // 获取文件夹树形结构
  tree: () => request.get<FolderTreeNode[]>(`${RESOURCE_PATH}/tree`),

  // 获取文件夹列表（分页）
  page: (page = 1, pageSize = 10, parentId?: string) =>
    request.get<FolderPageResponse>(`${RESOURCE_PATH}/page`, {
      params: { page, page_size: pageSize, parent_id: parentId },
    }),

  // 创建文件夹
  create: (data: CreateFolderRequest) => request.post<Folder>(RESOURCE_PATH, data),

  // 更新文件夹
  update: (data: UpdateFolderRequest) => request.put<Folder>(RESOURCE_PATH, data),

  // 删除文件夹
  delete: (id: string) => request.delete(`${RESOURCE_PATH}/${id}`),
};
