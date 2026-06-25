import { useState } from "react";

import {
  type ModelConfig,
  // useCreateModelConfig,
  useDeleteModelConfig,
  useInvalidateModelConfigList,
  useModelConfigList,
} from "@/api/ai/model-config";
import { ConfirmDialog } from "@/components/sag-ui/confirm-dialog";

import { ModelConfigForm } from "./_components/form";
import { ModelList } from "./_components/model-list";

export function ModelConfigPage() {
  // 获取模型配置列表（自动缓存、加载状态、错误处理）
  const { data: configs = [], isLoading } = useModelConfigList();

  // 选中的配置
  const [selectedConfig, setSelectedConfig] = useState<ModelConfig | null>(null);

  // 删除确认
  const [deleteTarget, setDeleteTarget] = useState<ModelConfig | null>(null);

  // 创建 mutation
  // const createMutation = useCreateModelConfig();

  // 删除 mutation
  const deleteMutation = useDeleteModelConfig();

  // 刷新（让 form 组件重新获取最新数据）
  const invalidateList = useInvalidateModelConfigList();

  // 删除配置
  // const handleDelete = (config: ModelConfig) => {
  //   setDeleteTarget(config);
  // };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;

    const wasSelected = selectedConfig?.id === deleteTarget.id;

    try {
      await deleteMutation.mutateAsync(deleteTarget.id);
    } finally {
      setDeleteTarget(null);
      // 如果删除的是选中的项，清除选中状态
      if (wasSelected) {
        setSelectedConfig(null);
      }
    }
  };

  // 选择配置
  const handleSelect = (config: ModelConfig | null) => {
    setSelectedConfig(config);
  };

  // 刷新列表
  const handleRefresh = () => {
    invalidateList();
  };

  return (
    <div className="h-screen p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">模型配置</h1>
        <p className="text-gray-600 dark:text-gray-400">管理 AI 模型配置和参数</p>
      </div>

      {/* 主要内容区域 */}
      <div className="flex gap-6 h-[calc(100vh-8rem)]">
        {/* 左侧模型列表 */}
        <div className="w-72 shrink-0">
          <div className="h-full overflow-auto">
            <ModelList
              configs={configs}
              selectedId={selectedConfig?.id}
              onSelect={handleSelect}
              loading={isLoading}
            />
          </div>
        </div>

        {/* 右侧配置表单 */}
        <div className="flex-1 min-w-0 overflow-auto">
          <ModelConfigForm
            key={selectedConfig?.id ?? "create"}
            config={selectedConfig}
            onRefresh={handleRefresh}
            onDelete={selectedConfig ? () => setDeleteTarget(selectedConfig) : undefined}
            updating={deleteMutation.isPending}
          />
        </div>
      </div>

      {/* 删除确认弹窗 */}
      <ConfirmDialog
        open={!!deleteTarget}
        title="删除配置"
        description={`确定要删除配置"${deleteTarget?.name}"吗？此操作不可撤销。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
