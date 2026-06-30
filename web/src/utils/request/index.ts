import type { AxiosRequestConfig, InternalAxiosRequestConfig } from "axios";
import axios from "axios";
import { toast } from "sonner";

import { useAuthStore } from "@/stores/auth";

export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
}

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

// Token 刷新队列
let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

function subscribeTokenRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb);
}

function onRefreshed(token: string) {
  refreshSubscribers.forEach((cb) => cb(token));
  refreshSubscribers = [];
}

async function refreshToken(): Promise<string | null> {
  const { refreshToken: refresh, setAccessToken } = useAuthStore.getState();
  if (!refresh) return null;

  try {
    const { data } = await axios.post<ApiResponse<{ access_token: string }>>(
      `${import.meta.env.VITE_API_BASE_URL}/auth/refresh-token`,
      { refresh_token: refresh },
    );
    if (data.code === 0 && data.data.access_token) {
      setAccessToken(data.data.access_token);
      return data.data.access_token;
    }
  } catch (error) {
    toast.error("刷新 token 失败", {
      description: error instanceof Error ? error.message : "未知错误",
    });
  }
  return null;
}

// 请求拦截器
instance.interceptors.request.use((config) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  // FormData 时删除 Content-Type，让浏览器自动处理 boundary
  if (config.data instanceof FormData) {
    delete config.headers["Content-Type"];
  }
  return config;
});

// 响应拦截器
instance.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (axios.isCancel(error)) {
      toast.error("请求已取消");
      return Promise.reject(error);
    }

    if (error.code === "ECONNABORTED") {
      toast.error("请求超时");
      return Promise.reject(new Error("请求超时"));
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    const status = error.response?.status;

    // 401 处理：尝试刷新 token
    if (
      status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes("/auth/login") &&
      !originalRequest.url?.includes("/auth/refresh-token")
    ) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh((newToken) => {
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            instance(originalRequest).then(resolve).catch(reject);
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const newToken = await refreshToken();
        if (newToken) {
          onRefreshed(newToken);
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return instance(originalRequest);
        }
        useAuthStore.getState().clearAuth();
        window.location.href = "/login";
        return Promise.reject(new Error("登录已过期"));
      } finally {
        isRefreshing = false;
      }
    }

    const msg = error.response?.data?.message || error.message || "请求失败";
    toast.error("请求失败", { description: msg });
    return Promise.reject(error);
  },
);

// 业务解包
async function unwrap<T>(
  promise: Promise<import("axios").AxiosResponse<ApiResponse<T>>>,
): Promise<T> {
  const res = await promise;
  const json = res.data;
  if (json.code === 0) return json.data;
  toast.error("操作失败", { description: json.message || "未知错误" });
  throw new Error(json.message || "未知错误");
}

// 便捷方法
export function request<T = unknown>(url: string, config?: AxiosRequestConfig) {
  return unwrap<T>(instance<ApiResponse<T>>(url, config));
}

request.get = <T = unknown>(url: string, config?: AxiosRequestConfig) =>
  unwrap<T>(instance.get<ApiResponse<T>>(url, config));

request.post = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
  unwrap<T>(instance.post<ApiResponse<T>>(url, data, config));

request.put = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
  unwrap<T>(instance.put<ApiResponse<T>>(url, data, config));

request.delete = <T = unknown>(url: string, config?: AxiosRequestConfig) =>
  unwrap<T>(instance.delete<ApiResponse<T>>(url, config));
