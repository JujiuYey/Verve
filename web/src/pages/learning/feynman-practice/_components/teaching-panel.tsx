import { GraduationCapIcon, NotebookPenIcon } from "lucide-react";

import { MessageResponse } from "@/components/ai-elements/message";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

export function TeachingPanel({
  tutorAdvice,
  isTutorTeaching,
  canRequestTeaching,
  canAppendTutorNote,
  isAppendingTutorNote,
  onRequestTutorTeaching,
  onAppendTutorNote,
}: {
  tutorAdvice: string;
  isTutorTeaching: boolean;
  canRequestTeaching: boolean;
  canAppendTutorNote: boolean;
  isAppendingTutorNote: boolean;
  onRequestTutorTeaching: () => void;
  onAppendTutorNote: () => void;
}) {
  return (
    <section className="flex min-h-0 flex-col overflow-hidden bg-background">
      <ScrollArea className="min-h-0 flex-1 px-4">
        <div className="flex min-h-full flex-col gap-4">
          <div className="flex shrink-0 items-start justify-between gap-3 rounded-lg bg-muted/30 px-3 py-2">
            <div className="flex items-start gap-2">
              <GraduationCapIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
              <p className="text-sm leading-6 text-muted-foreground">
                老师会根据刚才的复述结果补清楚关键点，再给出一段可以沉淀回 Markdown 的学习旁注。
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="shrink-0"
              onClick={onRequestTutorTeaching}
              disabled={!canRequestTeaching || isTutorTeaching}
            >
              <GraduationCapIcon data-icon="inline-start" />
              {isTutorTeaching ? "讲解中..." : tutorAdvice ? "重新讲一下" : "让老师讲解"}
            </Button>
          </div>

          <div className="rounded-lg border bg-background p-4">
            {tutorAdvice || isTutorTeaching ? (
              <MessageResponse className="max-w-none text-sm leading-6 text-muted-foreground">
                {tutorAdvice || "老师正在组织讲解..."}
              </MessageResponse>
            ) : (
              <div className="rounded-md border bg-background px-3 py-2">
                <div className="text-sm font-medium">
                  {canRequestTeaching ? "让老师补一段可沉淀的讲解" : "先完成一次复述判定"}
                </div>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">
                  {canRequestTeaching
                    ? "讲解会先补清楚知识点，再指出你漏掉的地方，最后给出可以写回 Markdown 的学习旁注。"
                    : "教学需要基于一次复述结果来补漏。先到复述页提交一次解释，再回到这里让老师讲。"}
                </p>
              </div>
            )}
          </div>

          <Button
            variant="secondary"
            size="sm"
            className="self-start"
            onClick={onAppendTutorNote}
            disabled={!canAppendTutorNote || !tutorAdvice.trim() || isAppendingTutorNote}
          >
            <NotebookPenIcon data-icon="inline-start" />
            {isAppendingTutorNote ? "追加中..." : "追加到 Markdown"}
          </Button>
        </div>
      </ScrollArea>
    </section>
  );
}