import { IconDatabase, IconPlus, IconTrash } from "@tabler/icons-react";
import { useState } from "react";

import {
  type Collection,
  useCollectionList,
  useCreateCollection,
  useDeleteCollection,
} from "@/api/ai/collection";
import { ConfirmDialog } from "@/components/sag-ui/confirm-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function CollectionPage() {
  const { data: collections = [], isLoading } = useCollectionList();
  const createMutation = useCreateCollection();
  const deleteMutation = useDeleteCollection();

  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Collection | null>(null);

  // 创建表单状态
  const [newName, setNewName] = useState("");
  const [newVectorSize, setNewVectorSize] = useState("1024");
  const [newDistance, setNewDistance] = useState("Cosine");

  const handleCreate = async () => {
    if (!newName.trim()) return;

    try {
      await createMutation.mutateAsync({
        name: newName.trim(),
        vector_size: parseInt(newVectorSize, 10),
        distance: newDistance,
      });
      setShowCreateDialog(false);
      setNewName("");
      setNewVectorSize("1024");
      setNewDistance("Cosine");
    } catch {
      // 错误已经在 mutation 中处理
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteMutation.mutateAsync(deleteTarget.name);
    } finally {
      setDeleteTarget(null);
    }
  };

  return (
    <div className="h-screen p-6 overflow-auto">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold mb-2">Qdrant Collection 管理</h1>
          <p className="text-gray-600 dark:text-gray-400">
            管理向量数据库中的 Collection，包括创建、删除和查看详情
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <IconPlus className="w-4 h-4 mr-2" />
          创建 Collection
        </Button>
      </div>

      {/* Collection 列表 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconDatabase className="w-5 h-5" />
            Collection 列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">加载中...</div>
          ) : collections.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              暂无 Collection，请点击上方按钮创建
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>向量维度</TableHead>
                  <TableHead>距离函数</TableHead>
                  <TableHead className="text-right">向量数量</TableHead>
                  <TableHead className="text-right">Points 数量</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {collections.map((col) => (
                  <TableRow key={col.name}>
                    <TableCell className="font-medium">{col.name}</TableCell>
                    <TableCell>
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          col.status === "Green"
                            ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300"
                            : "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
                        }`}
                      >
                        {col.status}
                      </span>
                    </TableCell>
                    <TableCell>{col.vector_size}</TableCell>
                    <TableCell>{col.distance_function}</TableCell>
                    <TableCell className="text-right">
                      {col.vectors_count.toLocaleString()}
                    </TableCell>
                    <TableCell className="text-right">
                      {col.points_count.toLocaleString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="destructive" size="sm" onClick={() => setDeleteTarget(col)}>
                        <IconTrash className="w-4 h-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* 创建 Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建 Collection</DialogTitle>
            <DialogDescription>
              创建一个新的向量 Collection，请确保向量维度与你的 embedding 模型匹配
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">名称</Label>
              <Input
                id="name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="例如: documents"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="vectorSize">向量维度</Label>
              <Select value={newVectorSize} onValueChange={setNewVectorSize}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="384">384 (轻量模型)</SelectItem>
                  <SelectItem value="512">512</SelectItem>
                  <SelectItem value="768">768 (BGE-base)</SelectItem>
                  <SelectItem value="1024">1024 (text-embedding-3-small, BGE-large)</SelectItem>
                  <SelectItem value="1536">1536 (text-embedding-ada-002)</SelectItem>
                  <SelectItem value="2048">2048</SelectItem>
                  <SelectItem value="3072">3072 (text-embedding-3-large)</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-gray-500">
                常见 embedding 模型维度：text-embedding-3-small=1024, text-embedding-3-large=3072,
                BGE-large=1024
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="distance">距离函数</Label>
              <Select value={newDistance} onValueChange={setNewDistance}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Cosine">Cosine (余弦相似度)</SelectItem>
                  <SelectItem value="Euclidean">Euclidean (欧几里得距离)</SelectItem>
                  <SelectItem value="Dot">Dot (点积)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              取消
            </Button>
            <Button onClick={handleCreate} disabled={!newName.trim() || createMutation.isPending}>
              {createMutation.isPending ? "创建中..." : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认 Dialog */}
      <ConfirmDialog
        open={!!deleteTarget}
        title="删除 Collection"
        description={`确定要删除 Collection "${deleteTarget?.name}" 吗？此操作不可撤销，所有数据将被永久删除。`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        onConfirm={handleDelete}
      />
    </div>
  );
}
