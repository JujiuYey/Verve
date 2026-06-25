import { ChevronRightIcon, UsersIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";

import type { Department } from "../_shared/types";

interface DepartmentRowProps {
  dept: Department;
  checked?: boolean;
  onCheck: () => void;
  onNavigate: () => void;
  hasChildren: boolean;
}

export function DepartmentRow({
  dept,
  checked,
  onCheck,
  onNavigate,
  hasChildren,
}: DepartmentRowProps) {
  return (
    <div className="flex items-center gap-2 px-2 py-2 hover:bg-accent rounded-md">
      <Checkbox checked={checked} onCheckedChange={onCheck} />
      <div className="size-7 rounded-full bg-primary/10 text-primary flex items-center justify-center shrink-0">
        <UsersIcon className="size-4" />
      </div>
      <span className="text-sm flex-1">{dept.name}</span>
      {hasChildren && (
        <Button
          variant="ghost"
          size="sm"
          className="h-auto px-1.5 py-0.5 text-xs gap-0.5 shrink-0"
          onClick={onNavigate}
        >
          下级
          <ChevronRightIcon className="size-3.5" />
        </Button>
      )}
    </div>
  );
}
