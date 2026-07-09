import z from "zod";

// 文件夹表单验证规则
export const folderFormSchema = z.object({
  name: z.string().min(1, "文件夹名称不能为空").max(100, "文件夹名称最多100个字符"),
  description: z.string().max(500, "描述最多500个字符").optional(),
  parent_id: z.string().optional(),
  sort_order: z
    .number()
    .int("排序值必须是整数")
    .min(0, "排序值不能小于0")
    .optional(),
});
