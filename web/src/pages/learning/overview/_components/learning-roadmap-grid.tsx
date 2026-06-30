import { Clock3Icon, LayersIcon, Trash2Icon, TrendingUpIcon, UsersIcon } from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { type LearningRoadmap } from "@/pages/learning/roadmap-adapter";

type Props = {
  roadmaps: LearningRoadmap[];
  onOpenRoadmap: (roadmap: LearningRoadmap) => void;
  onDeleteRoadmap: (roadmap: LearningRoadmap) => void;
  deletingRoadmapId?: string;
  isDeleting?: boolean;
};

export function LearningRoadmapGrid({
  roadmaps,
  onOpenRoadmap,
  onDeleteRoadmap,
  deletingRoadmapId,
  isDeleting = false,
}: Props) {
  return (
    <div className="grid grid-cols-1 gap-3 xl:grid-cols-2 2xl:grid-cols-3">
      {roadmaps.map((roadmap) => {
        const isDeletingThis = isDeleting && deletingRoadmapId === roadmap.id;
        return (
          <Card
            key={roadmap.id}
            className="rounded-2xl py-0 transition-colors hover:border-primary/50"
          >
            <CardHeader className="gap-3 border-b p-4!">
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-2.5">
                  <div className="space-y-1.5">
                    <CardTitle className="text-xl leading-7">{roadmap.title}</CardTitle>
                    <CardDescription className="leading-6">{roadmap.description}</CardDescription>
                  </div>
                </div>
                <Badge variant="outline">{roadmap.level}</Badge>
              </div>
            </CardHeader>

            <CardContent className="space-y-4">
              <p className="text-sm leading-6 text-muted-foreground">{roadmap.tagline}</p>

              <div className="grid grid-cols-2 gap-2.5">
                <InfoBlock icon={Clock3Icon} label="周期" value={roadmap.duration} />
                <InfoBlock icon={TrendingUpIcon} label="进度" value={`${roadmap.progress}%`} />
                <InfoBlock icon={LayersIcon} label="阶段数" value={`${roadmap.stages.length} 个`} />
                <InfoBlock icon={UsersIcon} label="学习人数" value={roadmap.learners} />
              </div>

              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">完成度</span>
                  <span className="font-medium">{roadmap.progress}%</span>
                </div>
                <Progress value={roadmap.progress} />
              </div>

              <div className="flex flex-wrap gap-2">
                {roadmap.tags.map((tag) => (
                  <Badge key={tag} variant="outline">
                    {tag}
                  </Badge>
                ))}
              </div>
            </CardContent>

            <CardFooter className="justify-between gap-2 border-t py-3">
              <DeleteRoadmapDialog
                roadmap={roadmap}
                disabled={isDeleting}
                isDeleting={isDeletingThis}
                onDelete={() => onDeleteRoadmap(roadmap)}
              />
              <Button onClick={() => onOpenRoadmap(roadmap)}>继续学习</Button>
            </CardFooter>
          </Card>
        );
      })}
    </div>
  );
}

function DeleteRoadmapDialog({
  roadmap,
  disabled,
  isDeleting,
  onDelete,
}: {
  roadmap: LearningRoadmap;
  disabled: boolean;
  isDeleting: boolean;
  onDelete: () => void;
}) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button variant="outline" disabled={disabled}>
          <Trash2Icon className="size-4" />
          删除
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除这个学习目标？</AlertDialogTitle>
          <AlertDialogDescription className="leading-6">
            {`将删除「${roadmap.title}」以及它下面的学习路径、小目标、练习会话、练习记录、学习日志和学习画像。文件管理里的文件夹和 Markdown 原文不会被删除。`}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>取消</AlertDialogCancel>
          <AlertDialogAction variant="destructive" disabled={isDeleting} onClick={onDelete}>
            {isDeleting ? "删除中..." : "确认删除"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function InfoBlock({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Clock3Icon;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-xl border bg-muted/30 px-3 py-2.5">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Icon className="size-3.5" />
        {label}
      </div>
      <div className="mt-0.5 text-sm font-medium">{value}</div>
    </div>
  );
}
