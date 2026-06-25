import { useForm } from "@tanstack/react-form";
import { useEffect } from "react";
import { z } from "zod";

import type {
  CreateDepartmentRequest,
  Department,
  UpdateDepartmentRequest,
} from "@/api/system/department";
import { TreeSelect, type TreeSelectItem } from "@/components/sag-ui/tree-select";
import { Button } from "@/components/ui/button";
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

import { departmentFormSchema } from "../_shared/form-schema";

export type DepartmentFormValues = z.infer<typeof departmentFormSchema>;

interface Props {
  open?: boolean;
  mode?: "create" | "edit";
  department?: Department | null;
  departmentTree?: Department[];
  loading?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSubmit?: (data: CreateDepartmentRequest | UpdateDepartmentRequest) => void;
}

// 部门树转 TreeSelectItem
function departmentToTreeSelectItems(
  departments: Department[],
  mode: "create" | "edit",
  currentDepartment?: Department | null,
): TreeSelectItem<Department>[] {
  return departments
    .filter((dept) => {
      if (mode === "edit" && currentDepartment) {
        if (dept.id === currentDepartment.id) return false;
        const hasChild = (parentId: string, nodes: Department[]): boolean => {
          for (const node of nodes) {
            if (node.id === parentId) return true;
            if (node.children && hasChild(parentId, node.children)) return true;
          }
          return false;
        };
        if (hasChild(dept.id, [currentDepartment])) return false;
      }
      return true;
    })
    .map((dept) => ({
      value: dept.id,
      label: dept.name,
      node: dept,
      children: dept.children
        ? departmentToTreeSelectItems(dept.children, mode, currentDepartment)
        : undefined,
    }));
}

export function DepartmentFormModal({
  open,
  mode = "create",
  department,
  departmentTree = [],
  loading = false,
  onOpenChange,
  onSubmit,
}: Props) {
  const form = useForm({
    defaultValues: {
      name: "",
      description: "",
      parent_id: undefined as string | undefined,
    } as DepartmentFormValues,
    validators: {
      onSubmit: departmentFormSchema,
    },
    onSubmit: ({ value }) => {
      if (!onSubmit) return;
      if (mode === "create") {
        onSubmit({
          name: value.name,
          description: value.description || undefined,
          parent_id: value.parent_id || undefined,
        });
      } else {
        onSubmit({
          id: department!.id,
          name: value.name,
          description: value.description || undefined,
          parent_id: value.parent_id || undefined,
        });
      }
    },
  });

  // dialog 打开时重置表单
  useEffect(() => {
    if (open) {
      form.reset();
      if (mode === "edit" && department) {
        form.setFieldValue("name", department.name);
        form.setFieldValue("description", department.description || "");
        form.setFieldValue("parent_id", department.parent_id);
      }
    }
  }, [open, mode, department, form]);

  // const departmentOptions = flattenDepartments(departmentTree, mode, department);
  const departmentTreeItems = departmentToTreeSelectItems(departmentTree, mode, department);
  console.log("🚀 ~ DepartmentFormModal ~ departmentTreeItems:", departmentTreeItems);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader className="gap-1">
          <DialogTitle>{mode === "create" ? "创建部门" : "编辑部门"}</DialogTitle>
          <DialogDescription>
            {mode === "create" ? "填写以下信息创建新部门" : "修改部门信息"}
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          className="flex flex-col gap-4"
        >
          <div className="flex flex-col gap-4 overflow-y-auto px-4 text-sm">
            {/* 部门名称 */}
            <form.Field name="name" validators={{ onChange: departmentFormSchema.shape.name }}>
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="name">
                    部门名称 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="name"
                    placeholder="请输入部门名称"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  {field.state.meta.isTouched && field.state.meta.errors.length > 0 && (
                    <p className="text-destructive text-xs">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            {/* 上级部门 */}
            <form.Field name="parent_id">
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="parent_id">上级部门</Label>
                  <TreeSelect
                    items={departmentTreeItems}
                    value={field.state.value}
                    onValueChange={(value) => field.handleChange(value || undefined)}
                    placeholder="请选择上级部门（可选）"
                    className="w-full"
                    allowClear
                    clearLabel="无上级部门"
                  />
                </div>
              )}
            </form.Field>

            {/* 描述 */}
            <form.Field
              name="description"
              validators={{ onChange: departmentFormSchema.shape.description }}
            >
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="description">描述</Label>
                  <Input
                    id="description"
                    placeholder="请输入部门描述（可选）"
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  {field.state.meta.isTouched && field.state.meta.errors.length > 0 && (
                    <p className="text-destructive text-xs">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={loading}>
              {loading ? "保存中..." : mode === "create" ? "创建" : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
