import { BookOpenTextIcon, FolderOpenIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "@tanstack/react-router";

export function FeynmanExercisePage() {
  return (
    <div className="flex h-full flex-col gap-4 overflow-auto p-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">费曼练习</h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
          学习项目和路线图已经移除。下一步会从 Wiki 文件夹进入学习现场，文件夹承载学习范围，文档承载每一次复述练习。
        </p>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <FolderOpenIcon className="size-4" />
            从 Wiki 文件夹开始
          </CardTitle>
          <CardDescription>
            现在先回到 Wiki 管理资料结构；后续会在文件夹上接入开始学习和继续学习。
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex items-start gap-3 rounded-lg border bg-muted/30 p-3 text-sm text-muted-foreground">
            <BookOpenTextIcon className="mt-0.5 size-4 shrink-0" />
            <div>
              已保留阅读、复述、讲解这个费曼小循环；只是旧的学习项目/学习路线入口已经下线。
            </div>
          </div>
          <Button asChild className="w-fit">
            <Link to="/wiki/folders">打开 Wiki</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
