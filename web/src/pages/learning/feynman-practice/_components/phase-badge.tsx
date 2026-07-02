import type { WorkbenchPhase } from "../_shared";

export function PhaseBadge({
  phase,
  onPhaseChange,
}: {
  phase: WorkbenchPhase;
  onPhaseChange: (phase: WorkbenchPhase) => void;
}) {
  return (
    <div className="flex items-center gap-1 rounded-md border bg-background p-1 text-xs">
      <button
        type="button"
        className={`rounded px-2 py-1 ${
          phase === "reading" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
        onClick={() => onPhaseChange("reading")}
      >
        1 阅读
      </button>
      <button
        type="button"
        className={`rounded px-2 py-1 ${
          phase === "answering" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
        onClick={() => onPhaseChange("answering")}
      >
        2 复述
      </button>
      <button
        type="button"
        className={`rounded px-2 py-1 ${
          phase === "teaching" ? "bg-primary text-primary-foreground" : "text-muted-foreground"
        }`}
        onClick={() => onPhaseChange("teaching")}
      >
        3 教学
      </button>
    </div>
  );
}
