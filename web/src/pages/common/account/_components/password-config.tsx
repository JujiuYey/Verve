import { IconDeviceFloppy, IconEye, IconEyeOff } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";

import { userApi } from "@/api/system/user";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth";

export function PasswordConfig() {
  const navigate = useNavigate();
  const clearAuth = useAuthStore((s) => s.clearAuth);

  const [formData, setFormData] = useState({
    oldPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [loading, setLoading] = useState(false);
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const resetForm = () => {
    setFormData({
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
    });
  };

  const handleSave = async () => {
    if (!formData.oldPassword) {
      toast.error("请输入原始密码");
      return;
    }

    if (!formData.newPassword) {
      toast.error("请输入新密码");
      return;
    }

    if (!formData.confirmPassword) {
      toast.error("请输入确认密码");
      return;
    }

    if (formData.newPassword.length < 6) {
      toast.error("密码长度至少为 6 位");
      return;
    }

    if (formData.newPassword !== formData.confirmPassword) {
      toast.error("两次输入的密码不一致");
      return;
    }

    if (formData.oldPassword === formData.newPassword) {
      toast.error("新密码不能与原密码相同");
      return;
    }

    setLoading(true);
    try {
      await userApi.changePassword({
        old_password: formData.oldPassword,
        new_password: formData.newPassword,
      });

      toast.success("密码修改成功，即将退出登录");
      resetForm();

      setTimeout(() => {
        clearAuth();
        navigate({ to: "/login" });
      }, 1500);
    } catch (error: any) {
      toast.error(error.message || "密码修改失败");
    } finally {
      setLoading(false);
    }
  };

  const renderPasswordField = (
    id: string,
    label: string,
    value: string,
    placeholder: string,
    show: boolean,
    onToggle: () => void,
    onChange: (value: string) => void,
    hint?: string,
  ) => (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="relative">
        <Input
          id={id}
          type={show ? "text" : "password"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="pr-10"
        />
        <button
          type="button"
          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          onClick={onToggle}
        >
          {show ? <IconEyeOff className="h-4 w-4" /> : <IconEye className="h-4 w-4" />}
        </button>
      </div>
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">密码设置</CardTitle>
        <CardDescription>修改您的登录密码</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {renderPasswordField(
          "oldPassword",
          "原始密码",
          formData.oldPassword,
          "请输入原始密码",
          showOldPassword,
          () => setShowOldPassword(!showOldPassword),
          (v) => setFormData((prev) => ({ ...prev, oldPassword: v })),
        )}

        {renderPasswordField(
          "newPassword",
          "新密码",
          formData.newPassword,
          "请输入新密码",
          showNewPassword,
          () => setShowNewPassword(!showNewPassword),
          (v) => setFormData((prev) => ({ ...prev, newPassword: v })),
          "密码长度至少为 6 位",
        )}

        {renderPasswordField(
          "confirmPassword",
          "确认密码",
          formData.confirmPassword,
          "请再次输入新密码",
          showConfirmPassword,
          () => setShowConfirmPassword(!showConfirmPassword),
          (v) => setFormData((prev) => ({ ...prev, confirmPassword: v })),
        )}

        {/* 保存按钮 */}
        <div className="flex justify-end pt-4">
          <Button disabled={loading} onClick={handleSave}>
            <IconDeviceFloppy className="mr-2 h-4 w-4" />
            {loading ? "保存中..." : "保存"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
