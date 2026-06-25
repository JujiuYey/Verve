import { UserIcon } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";

interface UserRowUser {
  id: string;
  name: string;
  avatar?: string;
}

interface UserRowProps {
  user: UserRowUser;
  checked?: boolean;
  onCheck: () => void;
  subtitle?: string;
}

export function UserRow({ user, checked, onCheck, subtitle }: UserRowProps) {
  return (
    <div className="flex items-center gap-2 px-2 py-2 hover:bg-accent rounded-md">
      <Checkbox checked={checked} onCheckedChange={onCheck} />
      <div className="size-7 rounded-full bg-muted flex items-center justify-center shrink-0">
        <UserIcon className="size-4 text-muted-foreground" />
      </div>
      <div className="flex flex-col min-w-0">
        <span className="text-sm">{user.name}</span>
        {subtitle && <span className="text-xs text-muted-foreground truncate">{subtitle}</span>}
      </div>
    </div>
  );
}
