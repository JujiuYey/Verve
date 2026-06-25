import { ChevronRight } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";

// 通用树节点接口
export interface TreeNode<T extends { id: string; children?: T[] }> {
  id: string;
  children?: T[];
}

interface TreeNodeComponentProps<T extends TreeNode<T>> {
  node: T;
  level: number;
  selectedId?: string;
  expandedIds: Set<string>;
  onSelect: (node: T) => void;
  onToggleExpand: (node: T) => void;
  renderNode: (node: T, isExpanded: boolean, isSelected: boolean) => React.ReactNode;
  hasChildren?: (node: T) => boolean;
}

function TreeNodeComponent<T extends TreeNode<T>>({
  node,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggleExpand,
  renderNode,
  hasChildren,
}: TreeNodeComponentProps<T>) {
  const isExpanded = expandedIds.has(node.id);
  const isSelected = selectedId === node.id;
  const nodeHasChildren = hasChildren
    ? hasChildren(node)
    : node.children && node.children.length > 0;

  const handleClick = () => {
    onSelect(node);
  };

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    onToggleExpand(node);
  };

  return (
    <div>
      <div
        className={cn(
          "group flex items-center gap-1 rounded-md px-2 py-1.5 cursor-pointer transition-colors",
          "hover:bg-muted/50",
          isSelected && "bg-muted",
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={handleClick}
      >
        {nodeHasChildren ? (
          <button
            onClick={handleToggle}
            className={cn(
              "p-0 border-none bg-transparent shrink-0 transition-transform",
              isExpanded && "rotate-90",
            )}
          >
            <ChevronRight className="size-4 text-muted-foreground" />
          </button>
        ) : (
          <span className="size-4 shrink-0" />
        )}

        <div className="flex items-center gap-2 flex-1">
          {renderNode(node, isExpanded, isSelected)}
        </div>
      </div>

      <Collapsible open={isExpanded}>
        <CollapsibleContent>
          {node.children?.map((child) => (
            <TreeNodeComponent
              key={child.id}
              node={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggleExpand={onToggleExpand}
              renderNode={renderNode}
              hasChildren={hasChildren}
            />
          ))}
        </CollapsibleContent>
      </Collapsible>
    </div>
  );
}

interface TreeProps<T extends TreeNode<T>> {
  data: T[];
  selectedId?: string;
  onSelect: (node: T | null) => void;
  renderNode: (node: T, isExpanded: boolean, isSelected: boolean) => React.ReactNode;
  hasChildren?: (node: T) => boolean;
  className?: string;
  rootLabel?: string;
  rootIcon?: React.ReactNode;
  loading?: boolean;
  defaultExpandedAll?: boolean;
}

// 递归收集所有节点 ID
function collectAllNodeIds<T extends TreeNode<T>>(nodes: T[]): Set<string> {
  const ids = new Set<string>();
  const collect = (nodeList: T[]) => {
    for (const node of nodeList) {
      ids.add(node.id);
      if (node.children && node.children.length > 0) {
        collect(node.children as T[]);
      }
    }
  };
  collect(nodes);
  return ids;
}

export function Tree<T extends TreeNode<T>>({
  data,
  selectedId,
  onSelect,
  renderNode,
  hasChildren,
  className,
  rootLabel,
  rootIcon,
  loading,
  defaultExpandedAll,
}: TreeProps<T>) {
  const initialExpandedIds = defaultExpandedAll ? collectAllNodeIds(data) : new Set<string>();
  const [expandedIds, setExpandedIds] = useState<Set<string>>(initialExpandedIds);

  // 当 defaultExpandedAll 为 true 且 data 更新时，重新展开所有节点
  useEffect(() => {
    if (defaultExpandedAll && data.length > 0) {
      setExpandedIds(collectAllNodeIds(data));
    }
  }, [defaultExpandedAll, data]);

  const handleToggleExpand = useCallback((node: T) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(node.id)) {
        next.delete(node.id);
      } else {
        next.add(node.id);
      }
      return next;
    });
  }, []);

  const handleSelectRoot = () => {
    onSelect(null);
  };

  return (
    <div className={cn("flex flex-col p-2", className)}>
      {rootLabel && (
        <div
          className={cn(
            "group flex items-center gap-2 rounded-md px-2 py-1.5 cursor-pointer transition-colors",
            "hover:bg-muted/50",
            !selectedId && "bg-muted",
          )}
          onClick={handleSelectRoot}
        >
          {rootIcon}
          <span className="text-sm">{rootLabel}</span>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-4">
          <span className="text-sm text-muted-foreground">加载中...</span>
        </div>
      ) : (
        data.map((node) => (
          <TreeNodeComponent
            key={node.id}
            node={node}
            level={0}
            selectedId={selectedId}
            expandedIds={expandedIds}
            onSelect={onSelect}
            onToggleExpand={handleToggleExpand}
            renderNode={renderNode}
            hasChildren={hasChildren}
          />
        ))
      )}
    </div>
  );
}
