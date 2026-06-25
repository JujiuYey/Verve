import { IconPlus, IconSearch } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import {
  type CreateDepartmentRequest,
  type Department,
  departmentApi,
  type UpdateDepartmentRequest,
} from "@/api/system/department";
import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import { DataTable } from "./_components/data-table";
import { DepartmentFormModal } from "./_components/department-form-modal";

export function DepartmentPage() {
  // 部门数据状态
  const [data, setData] = useState<Department[]>([]);
  const [departmentTree, setDepartmentTree] = useState<Department[]>([]);
  const [loading, setLoading] = useState(false);

  // 表单抽屉状态
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [selectedDepartment, setSelectedDepartment] = useState<Department | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [globalFilter, setGlobalFilter] = useState("");

  // 删除确认弹窗状态
  const [deleteTarget, setDeleteTarget] = useState<Department | null>(null);

  // 加载部门树
  const loadDepartmentTree = async () => {
    setLoading(true);
    try {
      const tree = await departmentApi.tree();
      const result = tree || [];
      setData(result);
      setDepartmentTree(result);
    } catch (error) {
      console.error("加载部门树失败:", error);
      toast.error("加载部门失败");
    } finally {
      setLoading(false);
    }
  };

  // 初始加载
  useEffect(() => {
    loadDepartmentTree();
  }, []);

  // 打开创建表单
  const handleCreate = () => {
    setFormMode("create");
    setSelectedDepartment(null);
    setFormOpen(true);
  };

  // 打开编辑表单
  const handleEdit = (department: Department) => {
    setFormMode("edit");
    setSelectedDepartment(department);
    setFormOpen(true);
  };

  // 删除部门
  const handleDelete = (department: Department) => {
    setDeleteTarget(department);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await departmentApi.delete(deleteTarget.id);
      toast.success("删除成功");
      loadDepartmentTree();
    } catch (error) {
      toast.error(`删除失败，${error}`);
    }
  };

  // 提交表单
  const handleSubmit = async (formData: CreateDepartmentRequest | UpdateDepartmentRequest) => {
    setSubmitting(true);
    try {
      if (formMode === "create") {
        await departmentApi.create(formData as CreateDepartmentRequest);
        toast.success("创建成功");
      } else {
        await departmentApi.update(formData as UpdateDepartmentRequest);
        toast.success("更新成功");
      }
      setFormOpen(false);
      loadDepartmentTree();
    } catch (error) {
      console.error("保存部门失败:", error);
      toast.error(formMode === "create" ? "创建失败" : "更新失败");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <div className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">部门管理</h1>
            <p className="text-muted-foreground mt-2">管理组织部门结构</p>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-4 px-6 pb-6">
        <div className="flex items-center justify-between gap-4">
          <div className="relative max-w-sm flex-1">
            <IconSearch className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索部门..."
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="pl-9"
            />
          </div>
          <Button onClick={handleCreate}>
            <IconPlus className="mr-2 h-4 w-4" />
            创建部门
          </Button>
        </div>
        <DataTable
          data={data}
          loading={loading}
          globalFilter={globalFilter}
          onGlobalFilterChange={setGlobalFilter}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </div>

      <DepartmentFormModal
        open={formOpen}
        mode={formMode}
        department={selectedDepartment}
        departmentTree={departmentTree}
        loading={submitting}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        title="删除部门"
        description={`确定要删除部门"${deleteTarget?.name}"吗？`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleConfirmDelete}
      />
    </>
  );
}
