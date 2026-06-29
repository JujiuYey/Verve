import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { useContinue, useCreateGoal, useGoalList } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

const EXAMPLES = ["Go 的并发", "Rust 所有权", "K8s 入门", "SQL 优化"];

export function LearningOverviewPage() {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");

  const { data: continueInfo } = useContinue();
  const { data: goalPage, isLoading } = useGoalList(1, 50);
  const createGoal = useCreateGoal();

  const goals = goalPage?.data ?? [];

  const handleCreate = async () => {
    const t = title.trim();
    if (!t) return;
    const res = await createGoal.mutateAsync({ title: t });
    setTitle("");
    if (res?.goal_id) {
      navigate({ to: "/learn/goal/$goalId", params: { goalId: res.goal_id } });
    }
  };

  return (
    <div className="flex h-full flex-col gap-6 overflow-auto p-6">
      {/* 新建目标 */}
      <Card className="p-6">
        <h2 className="mb-1 text-lg font-semibold">你想学什么?</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          一句话告诉我学习目标,AI 会自动拆出学习路线,并陪你一节节学。
        </p>
        <div className="flex gap-2">
          <Input
            placeholder="例如:我要学 Go 的并发"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleCreate();
            }}
            disabled={createGoal.isPending}
          />
          <Button onClick={handleCreate} disabled={createGoal.isPending || !title.trim()}>
            {createGoal.isPending ? "生成路线中…" : "生成学习路线"}
          </Button>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          {EXAMPLES.map((ex) => (
            <Badge
              key={ex}
              variant="secondary"
              className="cursor-pointer"
              onClick={() => setTitle(`我要学 ${ex}`)}
            >
              {ex}
            </Badge>
          ))}
        </div>
      </Card>

      {/* 继续上次 */}
      {continueInfo ? (
        <Card className="flex items-center justify-between p-4">
          <div>
            <div className="text-sm text-muted-foreground">继续上次</div>
            <div className="font-medium">{continueInfo.title ?? "上次的学习目标"}</div>
          </div>
          <Button
            onClick={() =>
              navigate({ to: "/learn/goal/$goalId", params: { goalId: continueInfo.goal_id } })
            }
          >
            继续 ▶
          </Button>
        </Card>
      ) : null}

      {/* 我的目标 */}
      <div>
        <h2 className="mb-3 text-lg font-semibold">我的学习目标</h2>
        {isLoading ? (
          <div className="text-sm text-muted-foreground">加载中…</div>
        ) : goals.length === 0 ? (
          <div className="text-sm text-muted-foreground">
            还没有学习目标,在上面创建第一个吧。
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {goals.map((goal) => (
              <Card
                key={goal.id}
                className="cursor-pointer p-4 transition-colors hover:border-primary"
                onClick={() => navigate({ to: "/learn/goal/$goalId", params: { goalId: goal.id } })}
              >
                <div className="mb-2 flex items-center justify-between gap-2">
                  <span className="font-medium">{goal.title}</span>
                  <Badge variant={goal.status === "completed" ? "default" : "secondary"}>
                    {goal.status}
                  </Badge>
                </div>
                <div className="text-xs text-muted-foreground">
                  {new Date(goal.created_at).toLocaleDateString()}
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
