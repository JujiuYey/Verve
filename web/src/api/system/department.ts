import { request } from "@/utils/request";

import type { User } from "../system/user";

// 部门接口定义
export interface Department {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  created_at: string;
  updated_at: string;
  children?: Department[];
  subRows?: Department[];
}

export interface CreateDepartmentRequest {
  name: string;
  description?: string;
  parent_id?: string;
}

export interface UpdateDepartmentRequest {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
}

export interface DepartmentListResponse {
  data: Department[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export interface Option {
  label: string;
  value: string;
}

export interface SearchResult {
  departments: Department[];
  users: User[];
}

const RESOURCE_PATH = "/api/system/department";

// 部门相关 API
export const departmentApi = {
  // 获取部门详情
  findOne: (id: string) => request.get<Department>(`${RESOURCE_PATH}/${id}`),

  // 获取树形结构的部门列表
  tree: () => request.get<Department[]>(`${RESOURCE_PATH}/tree`),

  // 创建部门
  create: (data: CreateDepartmentRequest) => request.post<Department>(RESOURCE_PATH, data),

  // 更新部门
  update: (data: UpdateDepartmentRequest) => request.put<Department>(`${RESOURCE_PATH}`, data),

  // 删除部门
  delete: (id: string) => request.delete(`${RESOURCE_PATH}/${id}`),

  // 获取子部门
  findChildren: (id: string) => request.get<Department[]>(`${RESOURCE_PATH}/${id}/children`),

  // 获取部门选项
  findOptions: () => request.get<Option[]>(`${RESOURCE_PATH}/options`),

  // 获取部门下的用户列表
  findUsers: (id: string) => request.get<User[]>(`${RESOURCE_PATH}/${id}/users`),

  // 搜索用户和部门
  search: (q: string) => request.get<SearchResult>(`${RESOURCE_PATH}/search`, { params: { q } }),
};
