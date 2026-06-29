import { useForm } from "@tanstack/react-form";
import { useEffect } from "react";

import type { CreateUserRequest, UpdateUserRequest, User } from "@/api/system/user";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { createUserFormSchema, updateUserFormSchema } from "../_shared/form-schema";

interface Props {
  open?: boolean;
  mode?: "create" | "edit";
  user?: User | null;
  loading?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSubmit?: (data: CreateUserRequest | UpdateUserRequest) => void;
}

function getErrorMessage(errors: unknown[]) {
  const firstError = errors[0];

  if (typeof firstError === "string") {
    return firstError;
  }

  if (
    firstError &&
    typeof firstError === "object" &&
    "message" in firstError &&
    typeof firstError.message === "string"
  ) {
    return firstError.message;
  }

  return undefined;
}

export function UserFormModal({
  open,
  mode = "create",
  user,
  loading = false,
  onOpenChange,
  onSubmit,
}: Props) {
  const isCreate = mode === "create";
  const schema = isCreate ? createUserFormSchema : updateUserFormSchema;

  const form = useForm({
    defaultValues: {
      username: "",
      email: "",
      password: "",
      full_name: "",
      status: "active",
    },
    validators: {
      onSubmit: schema as never,
    },
    onSubmit: ({ value }) => {
      if (!onSubmit) return;
      if (isCreate) {
        onSubmit({
          username: value.username,
          email: value.email,
          password: value.password || undefined,
          full_name: value.full_name || undefined,
        });
      } else {
        onSubmit({
          id: user!.id,
          email: value.email,
          full_name: value.full_name || undefined,
          status: value.status,
        });
      }
    },
  });

  useEffect(() => {
    if (open) {
      form.reset();
      if (mode === "edit" && user) {
        form.setFieldValue("email", user.email);
        form.setFieldValue("full_name", user.full_name || "");
        form.setFieldValue("status", user.status);
      }
    }
  }, [open, mode, user, form]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader className="gap-1">
          <DialogTitle>{isCreate ? "创建用户" : "编辑用户"}</DialogTitle>
          <DialogDescription>
            {isCreate ? "填写以下信息创建新用户" : "修改用户信息"}
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
            {/* 用户名 - 仅创建时 */}
            {isCreate && (
              <form.Field
                name="username"
                validators={{ onChange: createUserFormSchema.shape.username }}
              >
                {(field) => (
                  <div className="flex flex-col gap-3">
                    <Label htmlFor="username">
                      用户名 <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="username"
                      placeholder="请输入用户名"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                    {field.state.meta.isTouched && field.state.meta.errors.length > 0 ? (
                      <p className="text-destructive text-xs">
                        {getErrorMessage(field.state.meta.errors)}
                      </p>
                    ) : null}
                  </div>
                )}
              </form.Field>
            )}

            {/* 邮箱 */}
            <form.Field name="email" validators={{ onChange: schema.shape.email }}>
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="email">
                    邮箱 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="请输入邮箱"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  {field.state.meta.isTouched && field.state.meta.errors.length > 0 ? (
                    <p className="text-destructive text-xs">
                      {getErrorMessage(field.state.meta.errors)}
                    </p>
                  ) : null}
                </div>
              )}
            </form.Field>

            {/* 密码 - 仅创建时 */}
            {isCreate && (
              <form.Field
                name="password"
                validators={{ onChange: createUserFormSchema.shape.password as never }}
              >
                {(field) => (
                  <div className="flex flex-col gap-3">
                    <Label htmlFor="password">密码</Label>
                    <Input
                      id="password"
                      type="password"
                      placeholder="请输入密码（可选，留空则使用默认密码）"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                    {field.state.meta.isTouched && field.state.meta.errors.length > 0 ? (
                      <p className="text-destructive text-xs">
                        {getErrorMessage(field.state.meta.errors)}
                      </p>
                    ) : null}
                  </div>
                )}
              </form.Field>
            )}

            {/* 姓名 */}
            <form.Field name="full_name">
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="full_name">姓名</Label>
                  <Input
                    id="full_name"
                    placeholder="请输入姓名（可选）"
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            </form.Field>

            {/* 状态 - 仅编辑时 */}
            {!isCreate && (
              <form.Field name="status">
                {(field) => (
                  <div className="flex flex-col gap-3">
                    <Label htmlFor="status">
                      状态 <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={field.state.value}
                      onValueChange={(value) => field.handleChange(value)}
                    >
                      <SelectTrigger id="status" className="w-full">
                        <SelectValue placeholder="请选择状态" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="active">启用</SelectItem>
                        <SelectItem value="inactive">禁用</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
              </form.Field>
            )}
          </div>
          <DialogFooter>
            <Button type="submit" disabled={loading}>
              {loading ? "保存中..." : isCreate ? "创建" : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
