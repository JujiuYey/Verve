import z from "zod";

// 角色表单验证规则
export const roleFormSchema = z.object({
  name: z.string().min(1, "角色名称不能为空").max(50, "角色名称最多50个字符"),
  description: z.string().max(200, "描述最多200个字符").optional(),
});
