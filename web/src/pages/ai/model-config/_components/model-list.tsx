import { IconBrandOpenai, IconStar, IconStarFilled } from "@tabler/icons-react";

import { type ModelConfig } from "@/api/ai/model-config";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Spinner } from "@/components/ui/spinner";

interface ModelListProps {
  configs: ModelConfig[];
  selectedId?: string;
  onSelect: (config: ModelConfig | null) => void;
  loading?: boolean;
}

export function ModelList({ configs, selectedId, onSelect, loading }: ModelListProps) {
  return (
    <>
      <div className="space-y-4">
        {/* 标题和新建按钮 */}
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-muted-foreground">模型列表</h3>
          <Button
            size="sm"
            onClick={() => {
              onSelect(null);
            }}
          >
            创建配置
          </Button>
        </div>

        {/* 模型列表 */}
        <div className="space-y-1">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Spinner />
            </div>
          ) : configs.length === 0 ? (
            <Empty>
              <EmptyMedia variant="icon">
                <IconBrandOpenai className="h-6 w-6" />
              </EmptyMedia>
              <EmptyContent>
                <EmptyTitle>暂无配置</EmptyTitle>
                <EmptyDescription>创建第一个模型配置开始使用</EmptyDescription>
              </EmptyContent>
            </Empty>
          ) : (
            configs.map((config) => {
              const isSelected = selectedId === config.id;
              return (
                <div
                  key={config.id}
                  className={`group relative flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-all duration-200 ${
                    isSelected
                      ? "bg-primary text-primary-foreground shadow-sm"
                      : "hover:bg-accent hover:text-accent-foreground"
                  }`}
                  onClick={() => onSelect(config)}
                >
                  {/* 图标 */}
                  <div className={`shrink-0 ${isSelected ? "opacity-100" : "opacity-60"}`}>
                    <IconBrandOpenai className="h-5 w-5" />
                  </div>

                  {/* 名称和状态 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm truncate">{config.name}</span>
                      {config.is_default && (
                        <Badge
                          variant={isSelected ? "secondary" : "default"}
                          className={`shrink-0 ${isSelected ? "" : "bg-amber-500"}`}
                        >
                          {isSelected ? (
                            <IconStarFilled className="h-3 w-3" />
                          ) : (
                            <IconStar className="h-3 w-3" />
                          )}
                          默认
                        </Badge>
                      )}
                    </div>
                    <div
                      className={`text-xs truncate ${
                        isSelected ? "text-primary-foreground/70" : "text-muted-foreground"
                      }`}
                    >
                      {config.is_active ? "已启用" : "已禁用"}
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </>
  );
}
