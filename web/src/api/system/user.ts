import { request } from "@/utils/request";

import type { Department } from "./department";
import type { Role } from "./role";

// 用户接口定义
export interface User {
  id: string;
  username: string;
  email: string;
  full_name?: string;
  avatar?: string;
  status: string;
  created_at: string;
  updated_at: string;
  departments?: Department[];
  roles?: Role[];
  primary_department_id?: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password?: string;
  full_name?: string;
  avatar?: string;
  primary_department_id?: string;
  role_ids?: string[];
}

export interface UpdateUserRequest {
  id: string;
  email: string;
  full_name?: string;
  avatar?: string;
  status: string;
  department_ids?: string[];
  primary_department_id?: string;
  role_ids?: string[];
}

export interface ResetPasswordRequest {
  id: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface UpdateProfileRequest {
  full_name?: string;
  email: string;
  avatar?: string;
}

export interface UploadAvatarResponse {
  message: string;
  file_path: string;
}

export interface UserPageResponse {
  data: User[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

const RESOURCE_PATH = "/api/system/user";

// 用户相关 API
export const userApi = {
  // 获取用户详情
  findOne: (id: string) => request.get<User>(`${RESOURCE_PATH}/${id}`),

  // 获取用户列表
  page: (page = 1, pageSize = 10, search?: string) =>
    request.get<UserPageResponse>(`${RESOURCE_PATH}/page`, {
      params: { page, page_size: pageSize, search },
    }),

  // 创建用户
  create: (data: CreateUserRequest) => request.post<User>(RESOURCE_PATH, data),

  // 更新用户
  update: (data: UpdateUserRequest) => request.put<User>(RESOURCE_PATH, data),

  // 删除用户
  delete: (id: string) => request.delete(`${RESOURCE_PATH}/${id}`),

  // 重置密码
  resetPassword: (data: ResetPasswordRequest) =>
    request.post(`${RESOURCE_PATH}/reset-password`, data),

  // 上传头像
  uploadAvatar: (data: FormData) =>
    request.post<UploadAvatarResponse>(`${RESOURCE_PATH}/upload-avatar`, data),

  // 更新个人信息
  updateProfile: (data: UpdateProfileRequest) =>
    request.put(`${RESOURCE_PATH}/update-profile`, data),

  // 修改密码
  changePassword: (data: ChangePasswordRequest) =>
    request.put(`${RESOURCE_PATH}/change-password`, data),
};
