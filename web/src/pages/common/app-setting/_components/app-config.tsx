import { IconDeviceFloppy, IconRotate, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { useAppStore } from "@/stores/app";

export function AppConfig() {
  const settings = useAppStore((s) => s.settings);
  const updateSettings = useAppStore((s) => s.updateSettings);
  const resetSettings = useAppStore((s) => s.resetSettings);

  const [showClearConfirm, setShowClearConfirm] = useState(false);

  const handleSave = () => {
    updateSettings(settings);
    toast.success("设置已保存");
  };

  const handleReset = () => {
    resetSettings();
    toast.success("已恢复默认设置");
  };

  const handleClearAllConversations = () => {
    setShowClearConfirm(false);
    toast.success("所有对话已清除");
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">应用设置</CardTitle>
        <CardDescription>个性化应用体验</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 自动保存 */}
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label>自动保存对话</Label>
            <p className="text-sm text-muted-foreground">自动保存对话历史到本地存储</p>
          </div>
          <Switch
            checked={settings.autoSave}
            onCheckedChange={(checked) => updateSettings({ autoSave: checked })}
          />
        </div>

        <Separator />

        {/* 清除数据 */}
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label>清除数据</Label>
            <p className="text-sm text-muted-foreground">删除所有对话记录，此操作不可撤销</p>
          </div>
          <Button variant="destructive" size="sm" onClick={() => setShowClearConfirm(true)}>
            <IconTrash className="mr-2 h-4 w-4" />
            清除所有对话
          </Button>
        </div>

        <Separator />

        {/* 主题设置 */}
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label>主题设置</Label>
            <p className="text-sm text-muted-foreground">选择应用主题外观</p>
          </div>
          <Select
            value={settings.theme}
            onValueChange={(value: "system" | "light" | "dark") => updateSettings({ theme: value })}
          >
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="system">跟随系统</SelectItem>
              <SelectItem value="light">浅色</SelectItem>
              <SelectItem value="dark">深色</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>

      <CardFooter className="flex justify-between">
        <Button variant="outline" onClick={handleReset}>
          <IconRotate className="mr-2 h-4 w-4" />
          重置默认
        </Button>
        <Button onClick={handleSave}>
          <IconDeviceFloppy className="mr-2 h-4 w-4" />
          保存设置
        </Button>
      </CardFooter>

      {/* 清除确认对话框 */}
      <Dialog open={showClearConfirm} onOpenChange={setShowClearConfirm}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>确认清除</DialogTitle>
            <DialogDescription>
              你确定要清除所有对话记录吗？此操作不可撤销，所有聊天历史将被永久删除。
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={() => setShowClearConfirm(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleClearAllConversations}>
              <IconTrash className="mr-2 h-4 w-4" />
              确认清除
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
