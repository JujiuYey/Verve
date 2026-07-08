import { useForm } from "@tanstack/react-form";
import { useEffect } from "react";
import { z } from "zod";

import type { CreateFolderRequest, Folder, UpdateFolderRequest } from "@/api/wiki/folder";
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

import { folderFormSchema } from "../_shared/form-schema";

export type FolderFormValues = z.infer<typeof folderFormSchema>;

interface Props {
  open?: boolean;
  mode?: "create" | "edit";
  folder?: Folder | null;
  loading?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSubmit?: (data: CreateFolderRequest | UpdateFolderRequest) => void;
}

export function FolderFormModal({
  open,
  mode = "create",
  folder,
  loading = false,
  onOpenChange,
  onSubmit,
}: Props) {
  const form = useForm({
    defaultValues: {
      name: "",
      description: "",
      parent_id: undefined as string | undefined,
    } as FolderFormValues,
    validators: {
      onSubmit: folderFormSchema,
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
          id: folder!.id,
          name: value.name,
          description: value.description || undefined,
          parent_id: value.parent_id || undefined,
        });
      }
    },
  });

  useEffect(() => {
    if (open) {
      form.reset();
      if (mode === "edit" && folder) {
        form.setFieldValue("name", folder.name);
        form.setFieldValue("description", folder.description || "");
        form.setFieldValue("parent_id", folder.parent_id || undefined);
      }
    }
  }, [open, mode, folder, form]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader className="gap-1">
          <DialogTitle>{mode === "create" ? "添加文件夹" : "编辑文件夹"}</DialogTitle>
          <DialogDescription>
            {mode === "create" ? "创建新的文件夹" : "修改文件夹信息"}
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
            <form.Field name="name" validators={{ onChange: folderFormSchema.shape.name }}>
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="name">
                    文件夹名称 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="name"
                    placeholder="请输入文件夹名称"
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

            <form.Field
              name="description"
              validators={{ onChange: folderFormSchema.shape.description }}
            >
              {(field) => (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="description">描述</Label>
                  <Input
                    id="description"
                    placeholder="请输入文件夹描述（可选）"
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
            <Button type="button" variant="outline" onClick={() => onOpenChange?.(false)}>
              取消
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? "保存中..." : mode === "create" ? "创建" : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
