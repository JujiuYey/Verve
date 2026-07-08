import type { ReactNode } from "react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

interface ColorPickerPopoverProps {
  children: ReactNode;
  color: string | null;
  colors: string[];
  onChange: (color: string) => void;
}

export function ColorPickerPopover({ children, color, colors, onChange }: ColorPickerPopoverProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-auto p-2">
        <div className="flex max-w-40 flex-wrap gap-1">
          {colors.map((value) => (
            <button
              key={value}
              type="button"
              aria-label={value}
              className={cn(
                "h-6 w-6 rounded border border-border",
                color === value && "ring-2 ring-primary",
              )}
              style={{ backgroundColor: value }}
              onClick={() => onChange(value)}
            />
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}
