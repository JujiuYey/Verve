import { IconDeviceFloppy, IconUpload } from "@tabler/icons-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { userApi } from "@/api/system/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth";

export function AccountConfig() {
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);

  const [formData, setFormData] = useState({
    username: "",
    full_name: "",
    email: "",
    avatar: "",
  });
  const [loading, setLoading] = useState(false);
  const avatarInputRef = useRef<HTMLInputElement>(null);

  // 初始化表单
  useEffect(() => {
    if (user) {
      setFormData({
        username: user.username || "",
        full_name: user.full_name || "",
        email: user.email || "",
        avatar: user.avatar || "",
      });
    }
  }, [user]);

  const getUserInitial = () => {
    if (user?.full_name) {
      return user.full_name.charAt(0).toUpperCase();
    }
    return "U";
  };

  const avatarUrl = formData.avatar
    ? `${import.meta.env.VITE_API_BASE_URL}/api/files/${formData.avatar}`
    : "";

  const handleAvatarChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith("image/")) {
      toast.error("请选择图片文件");
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      toast.error("图片大小不能超过 5MB");
      return;
    }

    setLoading(true);
    try {
      const uploadFormData = new FormData();
      uploadFormData.append("file", file);
      const result = await userApi.uploadAvatar(uploadFormData);
      setFormData((prev) => ({ ...prev, avatar: result.file_path }));
      toast.success("头像上传成功");
    } catch (error: any) {
      toast.error(error.message || "头像上传失败");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!user) {
      toast.error("用户信息不存在");
      return;
    }

    if (!formData.email) {
      toast.error("邮箱不能为空");
      return;
    }

    setLoading(true);
    try {
      await userApi.updateProfile({
        email: formData.email,
        full_name: formData.full_name,
        avatar: formData.avatar,
      });

      // 更新本地用户信息
      setUser({
        ...user,
        email: formData.email,
        full_name: formData.full_name,
        avatar: formData.avatar,
      });

      toast.success("保存成功");
    } catch (error: any) {
      toast.error(error.message || "保存失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">账号信息</CardTitle>
        <CardDescription>管理您的个人信息</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* 头像上传 */}
        <div className="flex items-center gap-6">
          <Avatar className="h-20 w-20">
            <AvatarImage src={avatarUrl} alt={formData.full_name || user?.username} />
            <AvatarFallback className="text-2xl">{getUserInitial()}</AvatarFallback>
          </Avatar>
          <div className="space-y-2">
            <div>
              <Button variant="outline" size="sm" onClick={() => avatarInputRef.current?.click()}>
                <IconUpload className="mr-2 h-4 w-4" />
                上传头像
              </Button>
              <input
                ref={avatarInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleAvatarChange}
              />
            </div>
            <p className="text-xs text-muted-foreground">支持 JPG、PNG 格式，大小不超过 5MB</p>
          </div>
        </div>

        {/* 用户名（只读） */}
        <div className="space-y-2">
          <Label htmlFor="username">用户名</Label>
          <Input id="username" value={formData.username} disabled className="bg-muted" />
          <p className="text-xs text-muted-foreground">用户名不可修改</p>
        </div>

        {/* 姓名 */}
        <div className="space-y-2">
          <Label htmlFor="full_name">姓名</Label>
          <Input
            id="full_name"
            value={formData.full_name}
            onChange={(e) => setFormData((prev) => ({ ...prev, full_name: e.target.value }))}
            placeholder="请输入您的姓名"
          />
        </div>

        {/* 邮箱 */}
        <div className="space-y-2">
          <Label htmlFor="email">邮箱</Label>
          <Input
            id="email"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
            placeholder="请输入您的邮箱"
          />
        </div>

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
