import { request } from "@/utils/request";

// 角色接口定义
export interface Role {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
}

export interface UpdateRoleRequest {
  id: string;
  name: string;
  description?: string;
}

export interface RoleListResponse {
  data: Role[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const RESOURCE_PATH = "/api/system/role";

// 角色相关 API
export const roleApi = {
  // 获取角色详情
  findOne: (id: string) => request.get<Role>(`${RESOURCE_PATH}/${id}`),

  // 获取角色列表
  page: (page = 1, pageSize = 10) =>
    request.get<RoleListResponse>(`${RESOURCE_PATH}/page`, {
      params: { page, page_size: pageSize },
    }),

  // 创建角色
  create: (data: CreateRoleRequest) => request.post<Role>(RESOURCE_PATH, data),

  // 更新角色
  update: (data: UpdateRoleRequest) => request.put<Role>(`${RESOURCE_PATH}`, data),

  // 删除角色
  delete: (id: string) => request.delete(`${RESOURCE_PATH}/${id}`),
};
