import { IconDeviceFloppy, IconPlus, IconRefresh, IconStar, IconTrash } from "@tabler/icons-react";
import { useForm } from "@tanstack/react-form";
import { z } from "zod";

import type { ModelConfig, ModelType } from "@/api/ai/model-config";
import {
  useCreateModelConfig,
  useSetDefaultModelConfig,
  useUpdateModelConfig,
} from "@/api/ai/model-config";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";

import { CUSTOM_VENDOR_ID, getProvidersByType, MODEL_TYPES } from "../_shared/const";

interface ModelConfigFormProps {
  config: ModelConfig | null;
  onRefresh: () => void;
  onDelete?: () => void;
  updating?: boolean;
}

// 表单验证 schema
const formSchema = z.object({
  model_type: z.enum(["chat", "embedding"]),
  vendor: z.string().min(1, "请选择厂商"),
  name: z.string().min(1, "配置名称不能为空").max(50, "配置名称不能超过50个字符"),
  api_key: z.string().min(1, "API Key 不能为空"),
  base_url: z.string().min(1, "Base URL 不能为空"),
  model: z.string().min(1, "模型名称不能为空"),
  temperature: z.number().min(0, "Temperature 不能小于 0").max(2, "Temperature 不能大于 2"),
  top_p: z.number().min(0, "Top P 不能小于 0").max(1, "Top P 不能大于 1"),
  max_tokens: z
    .number()
    .int("Max Tokens 必须是整数")
    .positive("Max Tokens 必须是正数")
    .optional()
    .nullable(),
  top_k: z.number().int("Top K 必须是整数").positive("Top K 必须是正数").optional().nullable(),
  is_active: z.boolean(),
  is_default: z.boolean(),
});

const defaultValues = {
  model_type: "chat" as ModelType,
  vendor: "",
  name: "",
  api_key: "",
  base_url: "",
  model: "",
  temperature: 0.7,
  top_p: 0.9,
  max_tokens: null as number | null,
  top_k: null as number | null,
  is_active: true,
  is_default: false,
};

function getFormValues(config: ModelConfig | null) {
  if (!config) {
    return defaultValues;
  }

  return {
    model_type: config.model_type || "chat",
    vendor: config.vendor,
    name: config.name,
    api_key: config.api_key,
    base_url: config.base_url,
    model: config.model,
    temperature: config.temperature,
    top_p: config.top_p,
    max_tokens: config.max_tokens || null,
    top_k: config.top_k || null,
    is_active: config.is_active,
    is_default: config.is_default,
  };
}

// Helper to extract error message
function getErrorMessage(error: unknown): string {
  if (typeof error === "string") {
    return error;
  }
  if (error && typeof error === "object" && "message" in error) {
    return String(error.message);
  }
  return "";
}

export function ModelConfigForm({ config, onRefresh, onDelete }: ModelConfigFormProps) {
  const updateMutation = useUpdateModelConfig();
  const createMutation = useCreateModelConfig();
  const setDefaultMutation = useSetDefaultModelConfig();

  const form = useForm({
    defaultValues: getFormValues(config),
    onSubmit: async ({ value, formApi }) => {
      // 使用 zod 验证
      const result = formSchema.safeParse(value);
      if (!result.success) {
        // 设置字段错误
        const errors = result.error.flatten().formErrors;
        console.log("Validation errors:", errors);
        return;
      }

      // 准备数据，处理 null 为 undefined
      const data = {
        model_type: value.model_type,
        vendor: value.vendor,
        name: value.name,
        api_key: value.api_key,
        base_url: value.base_url,
        model: value.model,
        temperature: value.temperature,
        top_p: value.top_p,
        max_tokens: value.max_tokens || undefined,
        top_k: value.top_k || undefined,
        is_active: value.is_active,
        is_default: value.is_default,
      };

      if (config) {
        // 更新现有配置
        await updateMutation.mutateAsync({
          id: config.id,
          ...data,
        });
      } else {
        // 创建新配置
        await createMutation.mutateAsync(data);
      }

      // 刷新列表
      onRefresh();
      formApi.reset();
    },
  });

  const isLoading = updateMutation.isPending || createMutation.isPending;

  const isDefault = config ? (form.getFieldValue("is_default") ?? false) : false;

  // 获取当前模型类型对应的供应商
  const currentModelType = form.getFieldValue("model_type") || "chat";
  const providers = getProvidersByType(currentModelType);

  // 渲染错误消息
  const renderErrors = (errors: readonly unknown[]) => {
    if (errors.length === 0) {
      return null;
    }
    return (
      <p className="text-xs text-destructive">
        {errors
          .map((err) => getErrorMessage(err))
          .filter(Boolean)
          .join(", ")}
      </p>
    );
  };

  // 统一表单渲染函数
  const renderForm = (mode: "create" | "edit") => {
    const isEdit = mode === "edit";
    const isSubmitting = isEdit ? updateMutation.isPending : isLoading;
    const SubmitIcon = isEdit ? IconDeviceFloppy : IconPlus;
    const submitText = isEdit ? "保存配置" : "创建配置";

    const cardHeader = isEdit ? (
      <CardHeader>
        <CardTitle className="text-lg flex items-center gap-2">
          {isDefault && <IconStar className="h-4 w-4 text-amber-500" />}
          {config!.name}
        </CardTitle>
      </CardHeader>
    ) : (
      <CardHeader>
        <CardTitle>新建模型配置</CardTitle>
      </CardHeader>
    );

    return (
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void form.handleSubmit();
        }}
      >
        <Card>
          {cardHeader}

          <CardContent className="space-y-2">
            {/* 基本信息 */}
            <div className="grid gap-4 md:grid-cols-2">
              {/* 模型类型 */}
              <form.Field
                name="model_type"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || (value !== "chat" && value !== "embedding")) {
                      return "请选择模型类型";
                    }
                    return undefined;
                  },
                }}
                children={(field) => {
                  return (
                    <div className="space-y-2">
                      <Label htmlFor={field.name}>模型类型 *</Label>
                      <Select
                        value={field.state.value}
                        onValueChange={(value) => {
                          const newType = value as ModelType;
                          field.handleChange(newType);
                          // 切换类型时重置厂商和模型
                          form.setFieldValue("vendor", "");
                          form.setFieldValue("name", "");
                          form.setFieldValue("base_url", "");
                          form.setFieldValue("model", "");
                          // 更新供应商列表引用（用于模型选择）
                          if (newType === "embedding") {
                            form.setFieldValue("temperature", 0);
                            form.setFieldValue("top_p", 0.9);
                          } else {
                            form.setFieldValue("temperature", 0.7);
                            form.setFieldValue("top_p", 0.9);
                          }
                        }}
                      >
                        <SelectTrigger
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        >
                          <SelectValue placeholder="请选择模型类型" />
                        </SelectTrigger>
                        <SelectContent>
                          {MODEL_TYPES.map((type) => (
                            <SelectItem key={type.id} value={type.id}>
                              {type.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {renderErrors(field.state.meta.errors)}
                    </div>
                  );
                }}
              />

              {/* 厂商 */}
              <form.Field
                name="vendor"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || value.length === 0) {
                      return "请选择厂商";
                    }
                    return undefined;
                  },
                }}
                children={(field) => {
                  return (
                    <div className="space-y-2">
                      <Label htmlFor={field.name}>厂商 *</Label>
                      <Select
                        value={field.state.value}
                        onValueChange={(value) => {
                          const provider = providers.find((p) => p.id === value);
                          field.handleChange(value);
                          // 自动填充 name 和 base_url
                          if (provider && provider.id !== CUSTOM_VENDOR_ID) {
                            form.setFieldValue("name", provider.name);
                            form.setFieldValue("base_url", provider.base_url);
                            // 如果只有一个模型，直接填充
                            if (provider.models.length === 1) {
                              form.setFieldValue("model", provider.models[0]);
                            } else if (provider.models.length > 1) {
                              form.setFieldValue("model", "");
                            }
                          } else {
                            form.setFieldValue("name", "");
                            form.setFieldValue("base_url", "");
                            form.setFieldValue("model", "");
                          }
                        }}
                      >
                        <SelectTrigger
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        >
                          <SelectValue placeholder="请选择厂商" />
                        </SelectTrigger>
                        <SelectContent>
                          {providers.map((provider) => (
                            <SelectItem key={provider.id} value={provider.id}>
                              {provider.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {renderErrors(field.state.meta.errors)}
                    </div>
                  );
                }}
              />
            </div>

            {/* 配置名称 */}
            <form.Field
              name="name"
              validators={{
                onChange: ({ value }) => {
                  if (!value || value.length === 0) {
                    return "配置名称不能为空";
                  }
                  if (value.length > 50) {
                    return "配置名称不能超过50个字符";
                  }
                  return undefined;
                },
              }}
              children={(field) => {
                return (
                  <div className="space-y-2">
                    <Label htmlFor={field.name}>配置名称 *</Label>
                    <Input
                      id={field.name}
                      value={field.state.value}
                      onChange={(e) => {
                        field.handleChange(e.target.value);
                      }}
                      onBlur={field.handleBlur}
                      placeholder={
                        currentModelType === "embedding"
                          ? "例如：OpenAI Embedding"
                          : "例如：OpenAI GPT-4"
                      }
                      className={field.state.meta.errors.length ? "border-destructive" : ""}
                    />
                    {renderErrors(field.state.meta.errors)}
                  </div>
                );
              }}
            />

            {/* 模型名称 */}
            <form.Field
              name="model"
              validators={{
                onChange: ({ value }) => {
                  if (!value || value.length === 0) {
                    return "模型名称不能为空";
                  }
                  return undefined;
                },
              }}
              children={(field) => {
                const vendor = form.getFieldValue("vendor");
                const selectedProvider = providers.find((p) => p.id === vendor);
                const isCustom = vendor === CUSTOM_VENDOR_ID;
                const hasModels = selectedProvider && selectedProvider.models.length > 0;

                return (
                  <div className="space-y-2">
                    <Label htmlFor={field.name}>模型名称 *</Label>
                    {isCustom || !hasModels ? (
                      <Input
                        id={field.name}
                        value={field.state.value}
                        onChange={(e) => {
                          field.handleChange(e.target.value);
                        }}
                        onBlur={field.handleBlur}
                        placeholder={
                          currentModelType === "embedding"
                            ? "请输入自定义向量模型名称"
                            : "请输入自定义对话模型名称"
                        }
                        className={field.state.meta.errors.length ? "border-destructive" : ""}
                      />
                    ) : (
                      <Select value={field.state.value} onValueChange={field.handleChange}>
                        <SelectTrigger
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        >
                          <SelectValue placeholder="请选择模型" />
                        </SelectTrigger>
                        <SelectContent>
                          {selectedProvider.models.map((model) => (
                            <SelectItem key={model} value={model}>
                              {model}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                    {renderErrors(field.state.meta.errors)}
                  </div>
                );
              }}
            />

            {/* API 配置 */}
            <div className="space-y-4">
              {/* API Key */}
              <form.Field
                name="api_key"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || value.length === 0) {
                      return "API Key 不能为空";
                    }
                    return undefined;
                  },
                }}
                children={(field) => {
                  return (
                    <div className="space-y-2">
                      <Label htmlFor={field.name}>API Key *</Label>
                      <Input
                        id={field.name}
                        type="password"
                        value={field.state.value}
                        onChange={(e) => {
                          field.handleChange(e.target.value);
                        }}
                        onBlur={field.handleBlur}
                        placeholder="sk-xxxxxxxxxxxxxxxx"
                        className={field.state.meta.errors.length ? "border-destructive" : ""}
                      />
                      {renderErrors(field.state.meta.errors)}
                    </div>
                  );
                }}
              />

              {/* Base URL */}
              <form.Field
                name="base_url"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || value.length === 0) {
                      return "Base URL 不能为空";
                    }
                    return undefined;
                  },
                }}
                children={(field) => {
                  return (
                    <div className="space-y-2">
                      <Label htmlFor={field.name}>Base URL *</Label>
                      <Input
                        id={field.name}
                        value={field.state.value}
                        onChange={(e) => {
                          field.handleChange(e.target.value);
                        }}
                        onBlur={field.handleBlur}
                        placeholder="https://api.openai.com/v1"
                        className={field.state.meta.errors.length ? "border-destructive" : ""}
                      />
                      {renderErrors(field.state.meta.errors)}
                    </div>
                  );
                }}
              />
            </div>

            {/* Chat 模型参数 - embedding 模型隐藏 */}
            {currentModelType === "chat" && (
              <div className="grid gap-4 md:grid-cols-2">
                {/* Temperature */}
                <form.Field
                  name="temperature"
                  validators={{
                    onChange: ({ value }) => {
                      if (value < 0) {
                        return "Temperature 不能小于 0";
                      }
                      if (value > 2) {
                        return "Temperature 不能大于 2";
                      }
                      return undefined;
                    },
                  }}
                  children={(field) => {
                    return (
                      <div className="space-y-2">
                        <Label htmlFor={field.name}>
                          Temperature (随机性) *
                          <span className="text-xs text-muted-foreground ml-2">0-2</span>
                        </Label>
                        <Input
                          id={field.name}
                          type="number"
                          step="0.1"
                          min="0"
                          max="2"
                          value={field.state.value}
                          onChange={(e) => {
                            field.handleChange(parseFloat(e.target.value) || 0);
                          }}
                          onBlur={field.handleBlur}
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        />
                        <p className="text-xs text-muted-foreground">
                          控制输出的随机性，较低值更确定性，较高值更有创造性
                        </p>
                        {renderErrors(field.state.meta.errors)}
                      </div>
                    );
                  }}
                />

                {/* Top P */}
                <form.Field
                  name="top_p"
                  validators={{
                    onChange: ({ value }) => {
                      if (value < 0) {
                        return "Top P 不能小于 0";
                      }
                      if (value > 1) {
                        return "Top P 不能大于 1";
                      }
                      return undefined;
                    },
                  }}
                  children={(field) => {
                    return (
                      <div className="space-y-2">
                        <Label htmlFor={field.name}>
                          Top P (核采样) *
                          <span className="text-xs text-muted-foreground ml-2">0-1</span>
                        </Label>
                        <Input
                          id={field.name}
                          type="number"
                          step="0.1"
                          min="0"
                          max="1"
                          value={field.state.value}
                          onChange={(e) => {
                            field.handleChange(parseFloat(e.target.value) || 0);
                          }}
                          onBlur={field.handleBlur}
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        />
                        <p className="text-xs text-muted-foreground">
                          控制模型考虑的token范围，较低值更集中，较高值更多样
                        </p>
                        {renderErrors(field.state.meta.errors)}
                      </div>
                    );
                  }}
                />

                {/* Max Tokens */}
                <form.Field
                  name="max_tokens"
                  validators={{
                    onChange: ({ value }) => {
                      if (value === null || value === undefined) {
                        return undefined;
                      }
                      if (!Number.isInteger(value)) {
                        return "Max Tokens 必须是整数";
                      }
                      if (value <= 0) {
                        return "Max Tokens 必须是正数";
                      }
                      return undefined;
                    },
                  }}
                  children={(field) => {
                    return (
                      <div className="space-y-2">
                        <Label htmlFor={field.name}>
                          Max Tokens (最大Token数)
                          <span className="text-xs text-muted-foreground ml-2">可选</span>
                        </Label>
                        <Input
                          id={field.name}
                          type="number"
                          min="1"
                          value={field.state.value ?? ""}
                          onChange={(e) => {
                            field.handleChange(e.target.value ? parseInt(e.target.value) : null);
                          }}
                          onBlur={field.handleBlur}
                          placeholder="4096"
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        />
                        <p className="text-xs text-muted-foreground">生成回复的最大token数量限制</p>
                        {renderErrors(field.state.meta.errors)}
                      </div>
                    );
                  }}
                />

                {/* Top K */}
                <form.Field
                  name="top_k"
                  validators={{
                    onChange: ({ value }) => {
                      if (value === null || value === undefined) {
                        return undefined;
                      }
                      if (!Number.isInteger(value)) {
                        return "Top K 必须是整数";
                      }
                      if (value <= 0) {
                        return "Top K 必须是正数";
                      }
                      return undefined;
                    },
                  }}
                  children={(field) => {
                    return (
                      <div className="space-y-2">
                        <Label htmlFor={field.name}>
                          Top K (候选词数)
                          <span className="text-xs text-muted-foreground ml-2">可选</span>
                        </Label>
                        <Input
                          id={field.name}
                          type="number"
                          min="1"
                          value={field.state.value ?? ""}
                          onChange={(e) => {
                            field.handleChange(e.target.value ? parseInt(e.target.value) : null);
                          }}
                          onBlur={field.handleBlur}
                          placeholder="40"
                          className={field.state.meta.errors.length ? "border-destructive" : ""}
                        />
                        <p className="text-xs text-muted-foreground">
                          从最可能的token中采样时考虑的候选数量
                        </p>
                        {renderErrors(field.state.meta.errors)}
                      </div>
                    );
                  }}
                />
              </div>
            )}

            {/* 状态设置 */}
            <div className="grid gap-4 md:grid-cols-2">
              {/* 启用/禁用 */}
              <form.Field
                name="is_active"
                children={(field) => {
                  return (
                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label>启用配置</Label>
                        <p className="text-sm text-muted-foreground">
                          是否启用此配置，只有启用的配置才会被使用
                        </p>
                      </div>
                      <Switch checked={field.state.value} onCheckedChange={field.handleChange} />
                    </div>
                  );
                }}
              />

              {/* 设为默认 */}
              {/* <form.Field
                name="is_default"
                children={(field) => {
                  return (
                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label>设为默认</Label>
                        <p className="text-sm text-muted-foreground">
                          将此配置设为默认模型，当存在多个启用的配置时，默认配置优先使用
                        </p>
                      </div>
                      <Switch
                        checked={field.state.value}
                        onCheckedChange={field.handleChange}
                      />
                    </div>
                  );
                }}
              /> */}
            </div>
          </CardContent>

          <CardFooter className="flex justify-end gap-2">
            {config && onDelete ? (
              <>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setDefaultMutation.mutate(config.id)}
                  disabled={isDefault || setDefaultMutation.isPending}
                >
                  {setDefaultMutation.isPending ? (
                    <Spinner className="mr-2 h-4 w-4" />
                  ) : (
                    <IconStar className="mr-2 h-4 w-4" />
                  )}
                  {isDefault ? "已是默认" : "设为默认"}
                </Button>
                <Button type="button" variant="destructive" onClick={onDelete}>
                  <IconTrash className="mr-2 h-4 w-4" />
                  删除配置
                </Button>
              </>
            ) : (
              <Button type="button" variant="outline" onClick={() => form.reset()}>
                <IconRefresh className="mr-2 h-4 w-4" />
                重置
              </Button>
            )}
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <Spinner className="mr-2 h-4 w-4" />
              ) : (
                <SubmitIcon className="mr-2 h-4 w-4" />
              )}
              {submitText}
            </Button>
          </CardFooter>
        </Card>
      </form>
    );
  };

  return config ? renderForm("edit") : renderForm("create");
}
