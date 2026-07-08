import { cn } from "@/lib/utils";

interface ToolbarSelectProps {
  ariaLabel: string;
  className?: string;
  onValueChange: (value: string) => void;
  options: { value: string; label: string }[];
  value: string;
}

export function ToolbarSelect({
  ariaLabel,
  className,
  onValueChange,
  options,
  value,
}: ToolbarSelectProps) {
  return (
    <select
      aria-label={ariaLabel}
      className={cn(
        "h-8 rounded-md border border-input bg-background px-2 text-sm shadow-xs outline-none",
        className,
      )}
      value={value}
      onChange={(event) => onValueChange(event.target.value)}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
