import type { ModelType } from "@/api/system/model-config";

export const CUSTOM_VENDOR_ID = "custom";

export interface ProviderModel {
  id: string;
  name: string;
  group: string;
  type: ModelType;
  enabled: boolean;
}

export const MODEL_TYPES = [
  { id: "chat", name: "对话模型", description: "用于对话和生成文本" },
  { id: "embedding", name: "向量模型", description: "用于生成文本向量嵌入" },
  { id: "rerank", name: "重排模型", description: "用于检索结果重排" },
] as const;

export type ModelTypeId = (typeof MODEL_TYPES)[number]["id"];

const platformAccents: Record<string, string> = {
  dashscope: "from-emerald-500 to-teal-400",
  moonshot: "from-zinc-700 to-zinc-500",
  ark: "from-sky-500 to-cyan-400",
  deepseek: "from-blue-500 to-indigo-500",
  zhipu: "from-blue-500 to-indigo-400",
  silicon: "from-violet-500 to-indigo-400",
  minimax: "from-rose-500 to-pink-400",
  openai: "from-zinc-700 to-zinc-500",
  custom: "from-slate-500 to-slate-400",
};

export function getPlatformAccent(platformId: string) {
  return platformAccents[platformId] ?? platformAccents.custom;
}
