import { BookOpenCheckIcon, LightbulbIcon } from "lucide-react";

import type { LearningTeachingIntervention } from "@/api/learning";
import { Badge } from "@/components/ui/badge";

export function TeacherArtifact({ intervention }: { intervention: LearningTeachingIntervention }) {
  return (
    <div className="flex flex-col gap-4 text-sm">
      <div className="flex items-start gap-2">
        <LightbulbIcon className="mt-0.5 size-4 text-muted-foreground" />
        <div>
          <div className="font-medium">讲解重点</div>
          <p className="mt-1 text-muted-foreground">{intervention.explanation_summary}</p>
        </div>
      </div>
      {intervention.knowledge_gaps.length ? (
        <TextList title="需要补上的前置知识" items={intervention.knowledge_gaps} />
      ) : null}
      {intervention.key_points.length ? (
        <TextList title="关键点" items={intervention.key_points} />
      ) : null}
      {intervention.examples.length ? (
        <TextList title="例子" items={intervention.examples} />
      ) : null}
      {intervention.evidence.length ? (
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2 font-medium">
            <BookOpenCheckIcon className="size-4" />
            依据
          </div>
          <div className="flex flex-wrap gap-2">
            {intervention.evidence.map((evidence) => (
              <Badge key={`${evidence.chunk_id}-${evidence.chunk_index}`} variant="outline">
                v{evidence.document_version} ·{" "}
                {evidence.heading_path || `片段 ${evidence.chunk_index}`}
              </Badge>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}

function TextList({ title, items }: { title: string; items: string[] }) {
  return (
    <div>
      <div className="font-medium">{title}</div>
      <ul className="mt-1 flex list-disc flex-col gap-1 pl-5 text-muted-foreground">
        {items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </div>
  );
}
