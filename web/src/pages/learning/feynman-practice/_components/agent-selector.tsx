import { EarIcon, FilePenIcon, GraduationCapIcon } from "lucide-react";

import type { LearningAgentType } from "@/api/learning";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

const agents = [
  { value: "listener" as const, label: "倾听", icon: EarIcon },
  { value: "teacher" as const, label: "讲解", icon: GraduationCapIcon },
  { value: "curator" as const, label: "修订", icon: FilePenIcon },
];

export function AgentSelector({
  value,
  disabled,
  onChange,
}: {
  value: LearningAgentType;
  disabled: boolean;
  onChange: (value: LearningAgentType) => void;
}) {
  return (
    <ToggleGroup
      type="single"
      variant="outline"
      size="sm"
      value={value}
      disabled={disabled}
      onValueChange={(next) => next && onChange(next as LearningAgentType)}
      aria-label="选择学习 Agent"
    >
      {agents.map(({ value: agent, label, icon: Icon }) => (
        <ToggleGroupItem key={agent} value={agent} aria-label={label}>
          <Icon />
          {label}
        </ToggleGroupItem>
      ))}
    </ToggleGroup>
  );
}
