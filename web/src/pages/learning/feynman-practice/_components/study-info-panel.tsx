import type { ExerciseResult, LearningObjective } from "@/api/learning";
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
            {result ? (
              <div className="space-y-3 rounded-xl bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
                <InfoBlock label="判定依据" value={result.evidence} fallback="本次未返回判定依据" />
                <InfoBlock
                  label="薄弱点"
                  value={
                    result.weak_points && result.weak_points.length > 0
                      ? result.weak_points.join("、")
                      : "暂无明显薄弱点"
                  }
                />
                <InfoBlock
                  label="下一步"
                  value={result.next_recommendation}
                  fallback="继续推进下一个小目标"
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
                提交解释后，Learning Examiner 会同步写入练习记录、学习画像和本次日志。
              </div>
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
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
