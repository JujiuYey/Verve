import { ClipboardListIcon } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function ProfilePage() {
  return (
    <div className="flex h-full flex-col gap-4 overflow-auto p-6">
      <h1 className="text-2xl font-bold">我的画像</h1>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <ClipboardListIcon className="size-4" />
            等待接入 Wiki 文件夹学习现场
          </CardTitle>
          <CardDescription>
            学习项目和学习路线已经移除，画像后续会按 Wiki
            文件夹记录，像老师课前看的备课纸一样保存当前进度、已掌握内容和下一步。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-6 text-muted-foreground">
            旧的按学习目标选择画像入口已经下线。下一步从文件夹开始学习后，这里再展示对应文件夹下的学习状态。
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
