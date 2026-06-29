import z from "zod";

// 用户创建表单验证规则
export const createUserFormSchema = z.object({
  username: z.string().min(1, "用户名不能为空").max(50, "用户名最多50个字符"),
  email: z.string().min(1, "邮箱不能为空").email("请输入有效的邮箱地址"),
  password: z.string().min(6, "密码至少6个字符").max(50, "密码最多50个字符").optional(),
  full_name: z.string().max(50, "姓名最多50个字符").optional(),
});

// 用户编辑表单验证规则
export const updateUserFormSchema = z.object({
  email: z.string().min(1, "邮箱不能为空").email("请输入有效的邮箱地址"),
  full_name: z.string().max(50, "姓名最多50个字符").optional(),
  status: z.string().min(1, "状态不能为空"),
});
