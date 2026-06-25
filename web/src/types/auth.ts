// 认证相关类型定义

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: UserInfo;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
}

export interface UserInfo {
  id: string;
  username: string;
  email: string;
  full_name?: string;
  avatar?: string;
  status: string;
  roles: string[];
  primary_department_id?: string;
}
