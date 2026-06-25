import { IconPlus, IconSearch } from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import {
  type CreateRoleRequest,
  type Role,
  roleApi,
  type UpdateRoleRequest,
} from "@/api/system/role";
import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import { DataTable } from "./_components/data-table";
import { RoleFormModal } from "./_components/role-form-modal";

export function RolePage() {
  // 角色数据状态
  const [data, setData] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);

  // 表单弹窗状态
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState("");

  // 删除确认弹窗状态
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null);

  // 加载角色列表
  const loadRoles = useCallback(async () => {
    setLoading(true);
    try {
      const res = await roleApi.page(page, pageSize);
      setData(res.data || []);
      setTotal(res.total || 0);
    } catch (error) {
      toast.error(`加载角色列表失败，${error}`);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  // 页码或页大小变化时重新加载
  useEffect(() => {
    loadRoles();
  }, [loadRoles]);

  // 打开创建表单
  const handleCreate = () => {
    setFormMode("create");
    setSelectedRole(null);
    setFormOpen(true);
  };

  // 打开编辑表单
  const handleEdit = (role: Role) => {
    setFormMode("edit");
    setSelectedRole(role);
    setFormOpen(true);
  };

  // 删除角色
  const handleDelete = (role: Role) => {
    setDeleteTarget(role);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await roleApi.delete(deleteTarget.id);
      toast.success("删除成功");
      loadRoles();
    } catch (error) {
      toast.error(`删除失败，${error}`);
    }
  };

  // 提交表单
  const handleSubmit = async (formData: CreateRoleRequest | UpdateRoleRequest) => {
    setSubmitting(true);
    try {
      if (formMode === "create") {
        await roleApi.create(formData as CreateRoleRequest);
        toast.success("创建成功");
      } else {
        await roleApi.update(formData as UpdateRoleRequest);
        toast.success("更新成功");
      }
      setFormOpen(false);
      loadRoles();
    } catch (error) {
      console.error("保存角色失败:", error);
      toast.error(formMode === "create" ? "创建失败" : "更新失败");
    } finally {
      setSubmitting(false);
    }
  };

  // 页大小变化时重置到第一页
  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize);
    setPage(1);
  };

  // 过滤数据（前端搜索）
  const filteredData = searchKeyword
    ? data.filter(
        (role) =>
          role.name.toLowerCase().includes(searchKeyword.toLowerCase()) ||
          (role.description || "").toLowerCase().includes(searchKeyword.toLowerCase()),
      )
    : data;

  return (
    <>
      <div className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">角色管理</h1>
            <p className="text-muted-foreground mt-2">管理系统角色与权限</p>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-4 px-6 pb-6">
        <div className="flex items-center justify-between gap-4">
          <div className="relative max-w-sm flex-1">
            <IconSearch className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索角色..."
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              className="pl-9"
            />
          </div>
          <Button onClick={handleCreate}>
            <IconPlus className="mr-2 h-4 w-4" />
            创建角色
          </Button>
        </div>

        <DataTable
          data={filteredData}
          loading={loading}
          page={page}
          pageSize={pageSize}
          total={searchKeyword ? filteredData.length : total}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </div>

      <RoleFormModal
        open={formOpen}
        mode={formMode}
        role={selectedRole}
        loading={submitting}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        title="删除角色"
        description={`确定要删除角色"${deleteTarget?.name}"吗？`}
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
