import {
  useObjectiveExerciseList,
  type ExerciseResult,
  type LearningObjective,
} from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";

import { masteryLabels, verdictLabels } from "../_shared";

export function StudyInfoPanel({
  objective,
  result,
  sessionId,
}: {
  objective: LearningObjective;
  result: ExerciseResult | null;
  sessionId: string;
}) {
  const { data: history, isLoading: isHistoryLoading } = useObjectiveExerciseList(
    objective.id,
    1,
    3,
  );
  const historyItems = history?.data ?? [];
  const recentExercise = historyItems[0];
  const historyCount = history?.total ?? 0;

  return (
    <Card className="min-h-0 overflow-hidden rounded-2xl py-0">
      <CardHeader className="shrink-0 border-b p-4!">
        <CardTitle className="text-base">本次学习信息</CardTitle>
      </CardHeader>
      <CardContent className="min-h-0 flex-1 p-0">
        <ScrollArea className="h-full">
          <div className="space-y-4 p-4">
            <InfoRow label="会话状态" value={sessionId ? "已创建" : "创建中"} />
            <InfoRow label="小目标状态" value={objective.status} />
            <InfoRow
              label="原掌握度"
              value={masteryLabels[objective.mastery_level] ?? objective.mastery_level}
            />
            <Separator />
            <InfoRow
              label="本次判定"
              value={result ? (verdictLabels[result.verdict] ?? result.verdict) : "待提交"}
            />
            <InfoRow
              label="判定后掌握度"
              value={
                result
                  ? (masteryLabels[result.mastery_after] ?? result.mastery_after)
                  : (masteryLabels[objective.mastery_level] ?? objective.mastery_level)
              }
            />
            <Separator />
            <div className="space-y-3">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm font-medium">历史学习</div>
                <Badge variant="outline">
                  {isHistoryLoading ? "读取中" : `${historyCount} 次练习`}
                </Badge>
              </div>
              {recentExercise ? (
                <div className="space-y-3 rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                  <InfoRow
                    label="最近判定"
                    value={
                      recentExercise.verdict
                        ? (verdictLabels[recentExercise.verdict] ?? recentExercise.verdict)
                        : "未记录"
                    }
                  />
                  <InfoRow
                    label="最近掌握度"
                    value={
                      recentExercise.mastery_after
                        ? (masteryLabels[recentExercise.mastery_after] ??
                          recentExercise.mastery_after)
                        : "未记录"
                    }
                  />
                  <InfoRow label="最近练习" value={formatExerciseTime(recentExercise.created_at)} />
                  <InfoBlock
                    label="最近反馈"
                    value={recentExercise.feedback}
                    fallback="暂无历史反馈"
                  />
                </div>
              ) : (
                <div className="rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                  还没有这个小节的历史练习记录。提交一次解释后，这里会显示最近判定和反馈。
                </div>
              )}
            </div>
            {result ? (
              <div className="space-y-3 rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                <InfoBlock label="判定依据" value={result.evidence} fallback="本次未返回判定依据" />
                <InfoBlock
                  label="待补齐内容"
                  value={
                    result.weak_points && result.weak_points.length > 0
                      ? result.weak_points.join("、")
                      : "暂无需要补齐的内容"
                  }
                />
                <InfoBlock
                  label="改进建议"
                  value={result.improvement_suggestion ?? result.next_recommendation}
                  fallback="本次暂无额外改进建议"
                />
                <div className="flex items-center justify-between gap-3 border-t pt-3">
                  <span>复习标记</span>
                  <Badge variant={result.review_required ? "secondary" : "outline"}>
                    {result.review_required ? "需要复习" : "暂不需要"}
                  </Badge>
                </div>
              </div>
            ) : (
              <div className="rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                提交解释后，Learning Examiner 会同步写入练习记录、学习记忆和本次日志。
              </div>
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

function formatExerciseTime(value: string) {
  if (!value) return "未记录";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function InfoBlock({
  label,
  value,
  fallback = "-",
}: {
  label: string;
  value?: string;
  fallback?: string;
}) {
  return (
    <div className="space-y-1">
      <div className="text-xs font-medium text-foreground">{label}</div>
      <div className="break-words">{value || fallback}</div>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="shrink-0 text-muted-foreground">{label}</span>
      <span className="min-w-0 max-w-44 truncate text-right font-medium">{value}</span>
    </div>
  );
}
