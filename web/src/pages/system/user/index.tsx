import { IconPlus, IconSearch } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import {
  type CreateUserRequest,
  type UpdateUserRequest,
  type User,
  userApi,
} from "@/api/system/user";
import { ConfirmDialog } from "@/components/sag-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import { DataTable } from "./_components/data-table";
import { UserFormModal } from "./_components/user-form-modal";

export function UserPage() {
  // 用户数据状态
  const [data, setData] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState("");

  // 表单弹窗状态
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // 确认弹窗状态
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null);
  const [resetPasswordTarget, setResetPasswordTarget] = useState<User | null>(null);

  // 加载用户列表
  const loadUsers = async () => {
    setLoading(true);
    try {
      const res = await userApi.page(page, pageSize, searchKeyword || undefined);
      setData(res.data || []);
      setTotal(res.total || 0);
    } catch (error) {
      console.error("加载用户列表失败:", error);
      toast.error("加载用户列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, [page, pageSize]);

  // 搜索（回车或输入变化时重新加载）
  const handleSearch = () => {
    setPage(1);
    loadUsers();
  };

  // 打开创建表单
  const handleCreate = () => {
    setFormMode("create");
    setSelectedUser(null);
    setFormOpen(true);
  };

  // 打开编辑表单
  const handleEdit = (user: User) => {
    setFormMode("edit");
    setSelectedUser(user);
    setFormOpen(true);
  };

  // 删除用户
  const handleDelete = (user: User) => {
    setDeleteTarget(user);
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await userApi.delete(deleteTarget.id);
      toast.success("删除成功");
      loadUsers();
    } catch (error) {
      toast.error(`删除失败，${error}`);
    }
  };

  // 重置密码
  const handleResetPassword = (user: User) => {
    setResetPasswordTarget(user);
  };

  const handleConfirmResetPassword = async () => {
    if (!resetPasswordTarget) return;
    try {
      await userApi.resetPassword({ id: resetPasswordTarget.id });
      toast.success("密码已重置");
    } catch (error) {
      toast.error(`重置密码失败，${error}`);
    }
  };

  // 提交表单
  const handleSubmit = async (formData: CreateUserRequest | UpdateUserRequest) => {
    setSubmitting(true);
    try {
      if (formMode === "create") {
        await userApi.create(formData as CreateUserRequest);
        toast.success("创建成功");
      } else {
        await userApi.update(formData as UpdateUserRequest);
        toast.success("更新成功");
      }
      setFormOpen(false);
      loadUsers();
    } catch (error) {
      console.error("保存用户失败:", error);
      toast.error(formMode === "create" ? "创建失败" : "更新失败");
    } finally {
      setSubmitting(false);
    }
  };

  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize);
    setPage(1);
  };

  return (
    <>
      <div className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">用户管理</h1>
            <p className="text-muted-foreground mt-2">管理系统用户账号</p>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-4 px-6 pb-6">
        <div className="flex items-center justify-between gap-4">
          <div className="relative max-w-sm flex-1">
            <IconSearch className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索用户..."
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              className="pl-9"
            />
          </div>
          <Button onClick={handleCreate}>
            <IconPlus className="mr-2 h-4 w-4" />
            创建用户
          </Button>
        </div>
        <DataTable
          data={data}
          loading={loading}
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onResetPassword={handleResetPassword}
        />
      </div>

      <UserFormModal
        open={formOpen}
        mode={formMode}
        user={selectedUser}
        loading={submitting}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        title="删除用户"
        description={`确定要删除用户"${deleteTarget?.username}"吗？`}
        confirmText="删除"
        destructive
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleConfirmDelete}
      />

      <ConfirmDialog
        open={!!resetPasswordTarget}
        title="重置密码"
        description={`确定要重置用户"${resetPasswordTarget?.username}"的密码吗？`}
        confirmText="重置"
        onOpenChange={(open) => {
          if (!open) setResetPasswordTarget(null);
        }}
        onConfirm={handleConfirmResetPassword}
      />
    </>
  );
}
