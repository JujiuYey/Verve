import { useNavigate, useParams } from "@tanstack/react-router";

import { useCreateSession, useGoalDetail, type LearningObjective } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

export function GoalDetailPage() {
  const { goalId } = useParams({ from: "/_layout/learn/goal/$goalId" });
  const navigate = useNavigate();
  const { data, isLoading } = useGoalDetail(goalId);
  const createSession = useCreateSession();

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">加载中…</div>;
  }
  if (!data) {
    return <div className="p-6 text-sm text-muted-foreground">学习目标不存在</div>;
  }

  const { goal, objectives = [], current_objective_id, progress } = data;

  // 按阶段分组
  const stages: { title: string; items: LearningObjective[] }[] = [];
  for (const obj of objectives) {
    const stageTitle = obj.stage_title || "学习路线";
    let stage = stages.find((s) => s.title === stageTitle);
    if (!stage) {
      stage = { title: stageTitle, items: [] };
      stages.push(stage);
    }
    stage.items.push(obj);
  }

  const startSession = async (objectiveId: string) => {
    const res = await createSession.mutateAsync({ objective_id: objectiveId });
    if (res?.session_id) {
      navigate({ to: "/learn/session/$sessionId", params: { sessionId: res.session_id } });
    }
  };

  const pct =
    progress && progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0;

  return (
    <div className="flex h-full flex-col gap-6 overflow-auto p-6">
      <div>
        <div className="flex items-center justify-between gap-4">
          <h1 className="text-2xl font-bold">{goal.title}</h1>
          {current_objective_id ? (
            <Button
              onClick={() => startSession(current_objective_id)}
              disabled={createSession.isPending}
            >
              {createSession.isPending ? "进入中…" : "继续学习 ▶"}
            </Button>
          ) : null}
        </div>
        {progress ? (
          <div className="mt-3">
            <div className="mb-1 text-sm text-muted-foreground">
              进度 {progress.completed}/{progress.total}
            </div>
            <Progress value={pct} />
          </div>
        ) : null}
      </div>

      {stages.map((stage) => (
        <div key={stage.title}>
          <h2 className="mb-2 text-sm font-semibold text-muted-foreground">{stage.title}</h2>
          <div className="flex flex-col gap-2">
            {stage.items.map((obj) => {
              const isCurrent = obj.id === current_objective_id;
              const done = obj.status === "completed";
              return (
                <Card
                  key={obj.id}
                  className={`flex items-center justify-between gap-3 p-3 ${
                    isCurrent ? "border-primary" : ""
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-muted-foreground">
                      {done ? "✓" : isCurrent ? "●" : "○"}
                    </span>
                    <div>
                      <div className="font-medium">{obj.title}</div>
                      {obj.detail ? (
                        <div className="text-xs text-muted-foreground">{obj.detail}</div>
                      ) : null}
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <Badge variant="outline">{obj.mastery_level}</Badge>
                    <Button
                      size="sm"
                      variant={isCurrent ? "default" : "outline"}
                      onClick={() => startSession(obj.id)}
                      disabled={createSession.isPending}
                    >
                      {done ? "复习" : "开始"}
                    </Button>
                  </div>
                </Card>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
