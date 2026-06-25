import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";

import { authApi } from "@/api/auth/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/auth";

export function LoginForm() {
  const navigate = useNavigate();
  const { setTokens, setUser } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsLoading(true);

    const formData = new FormData(e.currentTarget);
    const username = formData.get("username") as string;
    const password = formData.get("password") as string;

    try {
      const response = await authApi.login({ username, password });
      setTokens(response.access_token, response.refresh_token);
      setUser(response.user);
      toast.success("登录成功");
      navigate({ to: "/" });
    } catch (error) {
      const message = error instanceof Error ? error.message : "登录失败";
      toast.error("登录失败", { description: message });
    } finally {
      setIsLoading(false);
    }
  };
  return (
    <div className="flex flex-col gap-6">
      <Card className="border-white/60 bg-white/88 shadow-[0_24px_80px_rgb(15_23_42/0.12)] backdrop-blur dark:border-white/10 dark:bg-card/92">
        <CardHeader className="space-y-1">
          <CardTitle>欢迎登录</CardTitle>
          <CardDescription>请输入账号密码，进入知识运营工作台</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="username">用户名</FieldLabel>
                <Input
                  id="username"
                  name="username"
                  type="text"
                  placeholder="请输入用户名"
                  required
                />
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">密码</FieldLabel>
                  <a
                    href="#"
                    className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                  >
                    忘记密码？
                  </a>
                </div>
                <Input
                  id="password"
                  name="password"
                  type="password"
                  required
                  placeholder="请输入密码"
                />
              </Field>
              <Field>
                <Button type="submit" disabled={isLoading}>
                  {isLoading ? "登录中..." : "登录"}
                </Button>
                <FieldDescription className="text-center">
                  使用统一账号登录，登录后可进入知识库、RAG 对话与系统管理。
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
