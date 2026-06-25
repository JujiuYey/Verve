import type {
  LoginRequest,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  UserInfo,
} from "@/types/auth";
import { request } from "@/utils/request";

const RESOURCE_PATH = "/api/auth";

// 认证相关 API
export const authApi = {
  // 登录
  login: (data: LoginRequest) => request.post<LoginResponse>(`${RESOURCE_PATH}/login`, data),

  // 刷新 token
  refreshToken: (data: RefreshTokenRequest) =>
    request.post<RefreshTokenResponse>(`${RESOURCE_PATH}/refresh`, data),

  // 获取当前用户信息
  getCurrentUser: () => request.get<UserInfo>(`${RESOURCE_PATH}/me`),
};
