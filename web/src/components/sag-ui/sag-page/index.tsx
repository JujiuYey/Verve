import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

interface SagPageProps {
  title: ReactNode;
  description?: ReactNode;
  icon?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  bodyClassName?: string;
}

export function SagPage({
  title,
  description,
  icon,
  actions,
  children,
  className,
  bodyClassName,
}: SagPageProps) {
  return (
    <div className={cn("flex min-h-0 flex-1 flex-col overflow-hidden p-6", className)}>
      <div className="mb-6 flex shrink-0 items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="mb-2 flex items-center gap-2 text-2xl font-bold">
            {icon}
            {title}
          </h1>
          {description ? <p className="text-muted-foreground">{description}</p> : null}
        </div>
        {actions ? <div className="shrink-0">{actions}</div> : null}
      </div>

      <div className={cn("min-h-0 flex-1 overflow-hidden", bodyClassName)}>{children}</div>
    </div>
  );
}
