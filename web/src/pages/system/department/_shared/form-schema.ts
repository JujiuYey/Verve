import z from "zod";

// 部门表单验证规则
export const departmentFormSchema = z.object({
  name: z.string().min(1, "部门名称不能为空").max(50, "部门名称最多50个字符"),
  description: z.string().max(200, "描述最多200个字符").optional(),
  parent_id: z.string().optional(),
});
