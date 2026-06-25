import type { MouseEventHandler, ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface ToolbarButtonProps {
  icon: ReactNode;
  title: string;
  onClick?: () => void;
  onDoubleClick?: MouseEventHandler<HTMLButtonElement>;
  disabled?: boolean;
  active?: boolean;
  isActive?: boolean;
  children?: ReactNode;
}

export function ToolbarButton({
  icon,
  title,
  onClick,
  onDoubleClick,
  disabled,
  active,
  isActive,
  children,
}: ToolbarButtonProps) {
  const buttonActive = active ?? isActive ?? false;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant={buttonActive ? "secondary" : "ghost"}
          size="icon-sm"
          title={title}
          aria-label={title}
          data-active={String(buttonActive)}
          onClick={onClick}
          onDoubleClick={onDoubleClick}
          disabled={disabled}
        >
          {icon}
          {children}
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        <p>{title}</p>
      </TooltipContent>
    </Tooltip>
  );
}
