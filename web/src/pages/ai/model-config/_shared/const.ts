export const CUSTOM_VENDOR_ID = "custom";

// 模型类型
export const MODEL_TYPES = [
  { id: "chat", name: "对话模型", description: "用于对话和生成文本" },
  { id: "embedding", name: "向量模型", description: "用于生成文本向量嵌入" },
] as const;

export type ModelTypeId = (typeof MODEL_TYPES)[number]["id"];

// Chat 模型供应商配置
export const CHAT_PROVIDERS = [
  {
    id: "openai",
    name: "OpenAI",
    base_url: "https://api.openai.com/v1",
    models: ["gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "o1", "o1-mini"],
  },
  {
    id: "deepseek",
    name: "DeepSeek",
    base_url: "https://api.deepseek.com/v1",
    models: ["deepseek-chat", "deepseek-reasoner"],
  },
  {
    id: "moonshot",
    name: "月之暗面",
    base_url: "https://api.moonshot.cn/v1",
    models: ["kimi-k2.5"],
  },
  {
    id: "dashscope-compatible",
    name: "通义千问",
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    models: ["qwen-plus", "qwen-turbo", "qwen-max", "qwen-long"],
  },
  {
    id: "zhipu",
    name: "智谱 GLM",
    base_url: "https://open.bigmodel.cn/api/paas/v4",
    models: ["glm-4-flash", "glm-4-plus", "glm-4-air", "glm-4-long"],
  },
  {
    id: "ark",
    name: "火山引擎",
    base_url: "https://ark.cn-beijing.volces.com/api/v3",
    models: ["doubao-seed-2-0-lite-260215"],
  },
  {
    id: "MiniMax",
    name: "MiniMax",
    base_url: "https://api.minimaxi.com/v1",
    models: ["MiniMax-M2.7", "MiniMax-M2.7-highspeed"],
  },
  {
    id: "custom",
    name: "自定义",
    base_url: "",
    models: [],
  },
];

// Embedding 模型供应商配置
export const EMBEDDING_PROVIDERS = [
  {
    id: "openai",
    name: "OpenAI",
    base_url: "https://api.openai.com/v1",
    models: ["text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002"],
  },
  {
    id: "ollama",
    name: "Ollama",
    base_url: "http://localhost:11434",
    models: ["bge-m3", "Qwen3-Embedding-4B"],
  },
  {
    id: "dashscope-compatible",
    name: "阿里云百炼",
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    models: ["text-embedding-v4"],
  },
  {
    id: "custom",
    name: "自定义",
    base_url: "",
    models: [],
  },
];

// 根据模型类型获取对应的供应商列表
export function getProvidersByType(modelType: ModelTypeId) {
  return modelType === "embedding" ? EMBEDDING_PROVIDERS : CHAT_PROVIDERS;
}
