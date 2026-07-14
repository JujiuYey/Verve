import type { AxiosRequestConfig } from "axios";
import axios from "axios";
import { toast } from "sonner";

export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
}

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

// 请求拦截器:不再注入 Authorization,后端已不再要求鉴权。
instance.interceptors.request.use((config) => {
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
