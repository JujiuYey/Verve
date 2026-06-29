import { useEffect, useMemo, useState } from "react";

import { useAIModels, useDeleteAIPlatform, useModelPlatforms } from "@/api/system/model-config";
import { ConfirmDialog, SagPage } from "@/components/sag-ui";
import { Spinner } from "@/components/ui/spinner";

import { CreatePlatformDialog } from "./_components/providers/create-platform-dialog";
import { ProviderList } from "./_components/providers/provider-list";
import { ProviderPanel } from "./_components/providers/provider-panel";

export function ModelConfigPage() {
  const platformsQuery = useModelPlatforms();
  const enabledModelsQuery = useAIModels();
  const deletePlatformMutation = useDeleteAIPlatform();
  const platforms = useMemo(() => platformsQuery.data ?? [], [platformsQuery.data]);
  const enabledModels = useMemo(() => enabledModelsQuery.data ?? [], [enabledModelsQuery.data]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [searchText, setSearchText] = useState("");
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  useEffect(() => {
    if (!selectedId && platforms[0]) {
      setSelectedId(platforms[0].id);
    }
  }, [platforms, selectedId]);

  const selectedPlatform = platforms.find((platform) => platform.id === selectedId) ?? null;
  const deleteTargetPlatform = platforms.find((platform) => platform.id === deleteTarget) ?? null;
  const loading = platformsQuery.isLoading || enabledModelsQuery.isLoading;
  const platformModels = selectedPlatform
    ? enabledModels.filter((model) => model.platform_id === selectedPlatform.id)
    : [];

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    await deletePlatformMutation.mutateAsync(deleteTarget);
    setDeleteTarget(null);
  };

  return (
    <SagPage
      title="模型配置"
      description="接入模型平台、保存平台密钥，并启用可用于内容生成与分析的模型。"
    >
      <div className="flex h-full min-h-0 flex-col overflow-hidden rounded-lg border bg-background">
        <div className="flex h-full min-h-0 overflow-hidden">
          <ProviderList
            platforms={platforms}
            enabledModels={enabledModels}
            selectedId={selectedId}
            search={searchText}
            onSearchChange={setSearchText}
            onSelect={setSelectedId}
            onAdd={() => setCreateDialogOpen(true)}
            onDelete={setDeleteTarget}
          />

          {loading ? (
            <div className="flex min-h-0 flex-1 items-center justify-center">
              <Spinner />
            </div>
          ) : selectedPlatform ? (
            <ProviderPanel platform={selectedPlatform} enabledModels={platformModels} />
          ) : (
            <div className="flex min-h-0 flex-1 items-center justify-center text-sm text-muted-foreground">
              暂无模型平台
            </div>
          )}
        </div>
      </div>

      <CreatePlatformDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={(platformId) => {
          setSelectedId(platformId);
        }}
      />

      <ConfirmDialog
        open={!!deleteTargetPlatform}
        title="删除模型平台"
        description={`确定要删除平台「${deleteTargetPlatform?.name}」吗？删除后无法恢复。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleConfirmDelete}
      />
    </SagPage>
  );
}
