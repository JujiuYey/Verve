import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useLearningMemory, type LearningMemoryItem } from "@/api/learning";

const KIND_LABELS: Record<string, string> = {
  mastered_concept: "已掌握概念",
  verification_evidence: "验证证据",
};

function getKindLabel(kind: string) {
  return KIND_LABELS[kind] ?? kind;
}

function formatSeenAt(value: string) {
  if (!value) return "";
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function MemoryItemCard({ item }: { item: LearningMemoryItem }) {
  return (
    <Card>
      <CardHeader className="gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{getKindLabel(item.kind)}</Badge>
          {item.confidence ? <Badge variant="outline">{item.confidence}</Badge> : null}
          {item.last_seen_at ? (
            <span className="text-xs text-muted-foreground">{formatSeenAt(item.last_seen_at)}</span>
          ) : null}
        </div>
        <CardTitle className="text-base leading-6">{item.statement}</CardTitle>
      </CardHeader>
    </Card>
  );
}

export function ProfilePage() {
  const { data: memory, isLoading } = useLearningMemory();
  const hasSummary = Boolean(memory?.summary?.trim());
  const hasItems = Boolean(memory?.items?.length);

  return (
    <div className="flex h-full flex-col gap-4 overflow-auto p-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">学习记忆</h1>
        <p className="max-w-3xl text-sm leading-6 text-muted-foreground">
          这里展示从回答、练习、文档和笔记中沉淀出来的证据型记忆，用来帮助后续学习更贴近你的真实掌握情况。
        </p>
      </div>

      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle className="text-base">记忆摘要</CardTitle>
          <CardDescription>基于最近学习证据生成的全局摘要。</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {isLoading ? (
            <p className="text-sm text-muted-foreground">正在读取学习记忆...</p>
          ) : hasSummary ? (
            <>
              <p className="text-sm leading-6">{memory?.summary}</p>
              {memory?.highlights?.length ? (
                <div className="flex flex-col gap-2">
                  {memory.highlights.map((highlight) => (
                    <p key={highlight} className="text-sm leading-6 text-muted-foreground">
                      {highlight}
                    </p>
                  ))}
                </div>
              ) : null}
            </>
          ) : (
            <p className="text-sm leading-6 text-muted-foreground">
              暂时还没有可展示的学习记忆。完成一次练习、记录笔记或基于文档学习后，这里会开始沉淀证据。
            </p>
          )}
        </CardContent>
      </Card>

      <div className="flex max-w-3xl flex-col gap-3">
        <h2 className="text-lg font-semibold">最近记忆条目</h2>
        {isLoading ? (
          <Card>
            <CardContent className="p-6 text-sm text-muted-foreground">正在读取最近条目...</CardContent>
          </Card>
        ) : hasItems ? (
          memory?.items.map((item) => <MemoryItemCard key={item.id} item={item} />)
        ) : (
          <Card>
            <CardContent className="p-6 text-sm leading-6 text-muted-foreground">
              还没有记忆条目。后续这里会列出已掌握概念、验证证据和其他学习观察。
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
