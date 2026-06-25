import { useForm } from "@tanstack/react-form";
import { useEffect } from "react";
import { z } from "zod";

import type { CreateRoleRequest, Role, UpdateRoleRequest } from "@/api/system/role";
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

import { roleFormSchema } from "../_shared/form-schema";

export type RoleFormValues = z.infer<typeof roleFormSchema>;

interface Props {
  open?: boolean;
  mode?: "create" | "edit";
  role?: Role | null;
  loading?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSubmit?: (data: CreateRoleRequest | UpdateRoleRequest) => void;
}

export function RoleFormModal({
  open,
  mode = "create",
  role,
  loading = false,
  onOpenChange,
  onSubmit,
}: Props) {
  const form = useForm({
    defaultValues: {
      name: "",
      description: "",
    } as RoleFormValues,
    validators: {
      onSubmit: roleFormSchema,
    },
    onSubmit: ({ value }) => {
      if (!onSubmit) return;
      if (mode === "create") {
        onSubmit({
          name: value.name,
          description: value.description || undefined,
        });
      } else {
        onSubmit({
          id: role!.id,
          name: value.name,
          description: value.description || undefined,
        });
      }
    },
  });

  // dialog 打开时重置表单
  useEffect(() => {
    if (open) {
      form.reset();
      if (mode === "edit" && role) {
        form.setFieldValue("name", role.name);
        form.setFieldValue("description", role.description || "");
      }
    }
  }, [open, mode, role, form]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader className="gap-1">
          <DialogTitle>{mode === "create" ? "创建角色" : "编辑角色"}</DialogTitle>
          <DialogDescription>
            {mode === "create" ? "填写以下信息创建新角色" : "修改角色信息"}
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
            {/* 角色名称 */}
            <form.Field name="name" validators={{ onChange: roleFormSchema.shape.name }}>
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="name">
                    角色名称 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="name"
                    placeholder="请输入角色名称"
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

            {/* 描述 */}
            <form.Field
              name="description"
              validators={{ onChange: roleFormSchema.shape.description }}
            >
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="description">描述</Label>
                  <Input
                    id="description"
                    placeholder="请输入角色描述（可选）"
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
