import { Folder as IconFolder, FolderOpen, Home } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import type { FolderTreeNode } from "@/api/wiki/folder";
import { folderApi } from "@/api/wiki/folder";
import { Tree } from "@/components/sag-ui/tree";

interface FolderTreeProps {
  data?: FolderTreeNode[];
  selectedId?: string;
  onSelect: (folder: FolderTreeNode | null) => void;
  className?: string;
}

export function FolderTree({ data, selectedId, onSelect, className }: FolderTreeProps) {
  const [internalTreeData, setInternalTreeData] = useState<FolderTreeNode[]>([]);
  const [loading, setLoading] = useState(false);

  const loadTree = useCallback(async () => {
    setLoading(true);
    try {
      const data = await folderApi.tree();
      setInternalTreeData(data || []);
    } catch (error) {
      console.error("加载文件夹树失败:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (data) return;
    void loadTree();
  }, [data, loadTree]);

  const treeData = data || internalTreeData;

  const renderFolderNode = (folder: FolderTreeNode, isExpanded: boolean, _isSelected: boolean) => {
    const Icon = isExpanded ? FolderOpen : IconFolder;
    return (
      <>
        <Icon className="size-4 text-primary shrink-0" />
        <span className="truncate text-sm">{folder.name}</span>
      </>
    );
  };

  return (
    <Tree
      data={treeData}
      selectedId={selectedId}
      onSelect={onSelect}
      renderNode={renderFolderNode}
      hasChildren={(folder) => folder.hasChildren}
      className={className}
      rootLabel="根目录"
      rootIcon={<Home className="size-4 text-primary shrink-0" />}
      loading={loading}
      defaultExpandedAll
    />
  );
}
