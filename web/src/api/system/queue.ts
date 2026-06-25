import { request } from "@/utils/request";

// 队列统计信息
export interface QueueStats {
  queue_name: string;
  pending: number;
  active: number;
  scheduled: number;
  retry: number;
  archived: number;
  completed: number;
}

// 任务信息
export interface TaskInfo {
  id: string;
  type: string;
  payload: Record<string, any>;
  queue: string;
  max_retry: number;
  retried: number;
  last_error?: string;
  state: string;
  next_process?: string;
}

// 任务列表响应
export interface TaskListResponse {
  tasks: TaskInfo[];
  total: number;
  queue_name: string;
  state: string;
  page: number;
  page_size: number;
}

const RESOURCE_PATH = "/api/system/queue";

// 队列相关 API
export const queueApi = {
  // 获取队列统计信息
  getStats: (queueName = "default") =>
    request.get<QueueStats>(`${RESOURCE_PATH}/stats`, {
      params: { queue: queueName },
    }),

  // 获取任务列表
  listTasks: (queueName = "default", state = "pending", pageSize = 20, page = 0) =>
    request.get<TaskListResponse>(`${RESOURCE_PATH}/tasks`, {
      params: {
        queue: queueName,
        state,
        page_size: pageSize,
        page,
      },
    }),
};
