import { SparklesIcon, TargetIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { useGoalList } from "@/api/learning";
import { useLearningProfile } from "@/api/learning/profile";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

function pickDefaultGoalId<T extends { id: string; status: string }>(
  goals: T[],
): string | undefined {
  const active = goals.find((g) => g.status === "active");
  if (active) return active.id;
  return goals[0]?.id;
}

export function ProfilePage() {
  const { data, isLoading, isError } = useGoalList(1, 100);
  const goals = useMemo(() => data?.data ?? [], [data?.data]);

  const [goalId, setGoalId] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (!goalId && goals.length > 0) {
      setGoalId(pickDefaultGoalId(goals));
    }
  }, [goalId, goals]);

  const { data: profile, isLoading: isProfileLoading } = useLearningProfile(goalId);

  const selectedGoal = useMemo(() => goals.find((g) => g.id === goalId), [goals, goalId]);

  if (isLoading) {
    return (
      <div className="flex h-full flex-col gap-4 overflow-auto p-6">
        <h1 className="text-2xl font-bold">我的画像</h1>
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    );
  }

  if (isError || goals.length === 0) {
    return (
      <div className="flex h-full flex-col gap-4 overflow-auto p-6">
        <h1 className="text-2xl font-bold">我的画像</h1>
        <Empty>
          <EmptyHeader>
            <EmptyTitle>{isError ? "加载学习目标失败" : "还没有学习目标"}</EmptyTitle>
            <EmptyDescription>
              {isError
                ? "请检查后端服务是否可用,稍后再试。"
                : "请先在「学习概览」中创建一个学习目标,目标完成后会自动生成画像。"}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-6 overflow-auto p-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">我的画像</h1>
        <p className="text-sm text-muted-foreground">
          按目标维度展示学习画像:当前水平、已掌握内容、验证习惯与下一步目标。
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-col gap-1">
              <CardTitle>选择学习目标</CardTitle>
              <CardDescription>
                {selectedGoal ? `当前查看:${selectedGoal.title}` : "请选择要查看画像的目标"}
              </CardDescription>
            </div>
            <Select value={goalId} onValueChange={setGoalId}>
              <SelectTrigger className="min-w-[240px]">
                <SelectValue placeholder="选择目标" />
              </SelectTrigger>
              <SelectContent>
                {goals.map((g) => (
                  <SelectItem key={g.id} value={g.id}>
                    {g.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
      </Card>

      {isProfileLoading ? (
        <div className="grid gap-4 md:grid-cols-2">
          <Skeleton className="h-40 rounded-xl" />
          <Skeleton className="h-40 rounded-xl" />
          <Skeleton className="h-32 rounded-xl md:col-span-2" />
        </div>
      ) : !profile ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>该目标尚未生成画像</EmptyTitle>
            <EmptyDescription>
              完成该目标的相关学习内容后,系统会自动生成画像,展示你已经掌握了什么。
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          <ProfileCurrentLevelCard level={profile.current_level} />
          <ProfileNextGoalCard nextGoal={profile.next_goal} />
          <ProfileCompletedTopicsCard topics={profile.completed_topics} />
          <ProfileVerificationHabitsCard habits={profile.verification_habits} />
        </div>
      )}

      <Separator />

      <p className="text-xs text-muted-foreground">
        画像在每次小目标验证后由系统更新,反映了该目标下的整体学习情况。
      </p>
    </div>
  );
}

function ProfileCurrentLevelCard({ level }: { level?: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <SparklesIcon className="size-4 text-primary" />
          当前水平
        </CardTitle>
        <CardDescription>综合验证后对该目标的整体掌握层级</CardDescription>
      </CardHeader>
      <CardContent>
        {level ? (
          <Badge variant="secondary" className="text-base">
            {level}
          </Badge>
        ) : (
          <p className="text-sm text-muted-foreground">尚未评估</p>
        )}
      </CardContent>
    </Card>
  );
}

function ProfileNextGoalCard({ nextGoal }: { nextGoal?: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TargetIcon className="size-4 text-primary" />
          下一步目标
        </CardTitle>
        <CardDescription>系统建议你接下来攻克的方向</CardDescription>
      </CardHeader>
      <CardContent>
        {nextGoal ? (
          <p className="text-sm leading-relaxed whitespace-pre-wrap">{nextGoal}</p>
        ) : (
          <p className="text-sm text-muted-foreground">暂未给出下一步建议</p>
        )}
      </CardContent>
    </Card>
  );
}

function ProfileCompletedTopicsCard({ topics }: { topics?: string[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>已掌握内容</CardTitle>
        <CardDescription>该目标下已经能够稳定复述和使用的知识点</CardDescription>
      </CardHeader>
      <CardContent>
        {topics && topics.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {topics.map((t) => (
              <Badge key={t} variant="secondary">
                {t}
              </Badge>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">还没有记录已掌握内容</p>
        )}
      </CardContent>
    </Card>
  );
}

function ProfileVerificationHabitsCard({ habits }: { habits?: string }) {
  return (
    <Card className="md:col-span-2">
      <CardHeader>
        <CardTitle>验证习惯</CardTitle>
        <CardDescription>你倾向用的验证方式与覆盖角度</CardDescription>
      </CardHeader>
      <CardContent>
        {habits ? (
          <p className="text-sm leading-relaxed whitespace-pre-wrap">{habits}</p>
        ) : (
          <p className="text-sm text-muted-foreground">还没有形成明显的验证习惯</p>
        )}
      </CardContent>
    </Card>
  );
}
