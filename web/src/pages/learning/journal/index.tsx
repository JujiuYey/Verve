import { useJournalList } from "@/api/learning";
import { Card } from "@/components/ui/card";

export function JournalPage() {
  const { data, isLoading } = useJournalList(1, 50);
  const journals = data?.data ?? [];

  return (
    <div className="flex h-full flex-col gap-4 overflow-auto p-6">
      <h1 className="text-2xl font-bold">学习日志</h1>
      {isLoading ? (
        <div className="text-sm text-muted-foreground">加载中…</div>
      ) : journals.length === 0 ? (
        <div className="text-sm text-muted-foreground">还没有学习日志,完成一节课后会自动生成。</div>
      ) : (
        <div className="flex flex-col gap-3">
          {journals.map((j) => (
            <Card key={j.id} className="p-4">
              <div className="mb-1 text-sm font-medium">
                {new Date(j.date).toLocaleDateString()}
              </div>
              {j.learned ? <div className="text-sm">学了:{j.learned}</div> : null}
              {j.weak_points ? (
                <div className="text-sm text-muted-foreground">待补齐:{j.weak_points}</div>
              ) : null}
              {j.next_step ? (
                <div className="text-sm text-muted-foreground">改进建议:{j.next_step}</div>
              ) : null}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
