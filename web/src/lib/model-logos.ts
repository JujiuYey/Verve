import AihubmixLogo from "@/assets/model-config/aihubmix.webp";
import BailianLogo from "@/assets/model-config/bailian-color.svg";
import DeepseekLogo from "@/assets/model-config/deepseek.svg";
import DmxapiLogo from "@/assets/model-config/dmxapi.png";
import DoubaoAvatar from "@/assets/model-config/doubao-avatar.png";
import DoubaoLogo from "@/assets/model-config/doubao.svg";
import EmbeddingLogo from "@/assets/model-config/embedding.png";
import KimiLogo from "@/assets/model-config/kimi.svg";
import MinimaxLogo from "@/assets/model-config/minimax.svg";
import OpenaiProviderLogo from "@/assets/model-config/openai.svg";
import QwenLogo from "@/assets/model-config/qwen.svg";
import SiliconLogo from "@/assets/model-config/silicon.png";
import VolcengineLogo from "@/assets/model-config/volcengine.png";
import WenxinLogo from "@/assets/model-config/wenxin.svg";
import YuanbaoLogo from "@/assets/model-config/yuanbao.svg";
import ZhipuLogo from "@/assets/model-config/zhipu.png";
import ZhipuProviderLogo from "@/assets/model-config/zhipu.png";

type ProviderLogoTarget = {
  id?: string;
  name?: string;
  provider_type?: string;
  default_base_url?: string;
};

const providerLogoRules: Array<[RegExp, string]> = [
  [/openai|api\.openai\.com/i, OpenaiProviderLogo],
  [/deepseek|api\.deepseek\.com/i, DeepseekLogo],
  [/kimi|moonshot|api\.moonshot\.cn/i, KimiLogo],
  [/阿里云?百炼|bailian|dashscope|aliyun|alibaba/i, BailianLogo],
  [/火山方舟|volcengine|volces|ark/i, VolcengineLogo],
  [/minimax|minimaxi/i, MinimaxLogo],
  [/zhipu|智谱|bigmodel/i, ZhipuProviderLogo],
  [/silicon|硅基流动/i, SiliconLogo],
  [/aihubmix/i, AihubmixLogo],
  [/dmxapi/i, DmxapiLogo],
];

const modelLogoRules: Array<[RegExp, string]> = [
  [/deepseek/i, DeepseekLogo],
  [/doubao|seed/i, DoubaoLogo],
  [/minimax/i, MinimaxLogo],
  [/gpt|chatgpt|o1|o3|o4|text-embedding-3/i, OpenaiProviderLogo],
  [/qwen|qwq|qvq|tongyi|千问/i, QwenLogo],
  [/glm|zhipu|chatglm/i, ZhipuLogo],
  [/embedding|embed|bge|gte|m3e|jina|rerank|reranker|embo/i, EmbeddingLogo],
  [/kimi/i, KimiLogo],
  [/豆包/i, DoubaoAvatar],
  [/文心/i, WenxinLogo],
  [/元宝/i, YuanbaoLogo],
];

// 获取平台 Logo
export function getProviderLogo(platform: ProviderLogoTarget) {
  const searchable = [platform.name, platform.default_base_url].filter(Boolean).join(" ");

  return providerLogoRules.find(([pattern]) => pattern.test(searchable))?.[1];
}

// 获取模型 Logo
export function getModelLogo(modelId: string) {
  return modelLogoRules.find(([pattern]) => pattern.test(modelId))?.[1];
}
