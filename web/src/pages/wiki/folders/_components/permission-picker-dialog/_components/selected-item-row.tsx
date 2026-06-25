import { UserIcon, UsersIcon, XIcon } from "lucide-react";

import { Button } from "@/components/ui/button";

import type { SelectedItem } from "../_shared/types";

interface SelectedItemRowProps {
  item: SelectedItem;
  onRemove: (id: string) => void;
}

export function SelectedItemRow({ item, onRemove }: SelectedItemRowProps) {
  return (
    <div className="flex items-center justify-between py-1.5 group">
      <div className="flex items-center gap-2 min-w-0">
        {item.type === "department" ? (
          <div className="size-6 rounded-full bg-primary/10 text-primary flex items-center justify-center shrink-0">
            <UsersIcon className="size-3.5" />
          </div>
        ) : (
          <div className="size-6 rounded-full bg-muted flex items-center justify-center shrink-0">
            <UserIcon className="size-3.5 text-muted-foreground" />
          </div>
        )}
        <span className="text-sm truncate">{item.name}</span>
      </div>
      <Button variant="destructive" size="icon-xs" onClick={() => onRemove(item.id)}>
        <XIcon className="size-3.5" />
      </Button>
    </div>
  );
}
