"use client";

import { CheckIcon, ChevronDown, ChevronRight } from "lucide-react";
import * as React from "react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export interface TreeSelectNode<T extends TreeSelectNode<T>> {
  id: string;
  children?: T[];
}

export interface TreeSelectItem<T extends TreeSelectNode<T>> {
  value: string;
  label: string;
  node: T;
  children?: TreeSelectItem<T>[];
}

export interface TreeSelectProps<T extends TreeSelectNode<T>> {
  items: TreeSelectItem<T>[];
  value?: string;
  onValueChange: (value: string, node: T | null) => void;
  placeholder?: string;
  allowClear?: boolean;
  clearLabel?: string;
  disabled?: boolean;
  className?: string;
  size?: "sm" | "default";
  emptyMessage?: string;
}

interface TreeSelectBranchProps<T extends TreeSelectNode<T>> {
  items: TreeSelectItem<T>[];
  level?: number;
  selectedValue?: string;
  onSelect: (value: string, node: T) => void;
  expandedIds: Set<string>;
  onToggleExpand: (id: string) => void;
}

function TreeSelectBranch<T extends TreeSelectNode<T>>({
  items,
  level = 0,
  selectedValue,
  onSelect,
  expandedIds,
  onToggleExpand,
}: TreeSelectBranchProps<T>) {
  return (
    <>
      {items.map((item) => {
        const hasChildren = item.children && item.children.length > 0;
        const isExpanded = expandedIds.has(item.value);
        const isSelected = selectedValue === item.value;

        return (
          <div key={item.value}>
            <div
              role="button"
              tabIndex={0}
              aria-pressed={isSelected}
              className={cn(
                "relative flex items-center gap-1 py-1.5 pr-8 text-sm rounded-sm cursor-pointer",
                "hover:bg-accent hover:text-accent-foreground",
                "outline-none",
                "focus-visible:bg-accent focus-visible:text-accent-foreground",
                isSelected && "bg-accent text-accent-foreground",
              )}
              style={{ paddingLeft: `${level * 16 + 8}px` }}
              onClick={() => onSelect(item.value, item.node)}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onSelect(item.value, item.node);
                }
              }}
            >
              {hasChildren ? (
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    onToggleExpand(item.value);
                  }}
                  className="p-0.5 hover:bg-accent rounded-sm"
                >
                  {isExpanded ? (
                    <ChevronDown className="size-3.5" />
                  ) : (
                    <ChevronRight className="size-3.5" />
                  )}
                </button>
              ) : (
                <span className="size-3.5" />
              )}
              <span className="truncate">{item.label}</span>
              {isSelected && (
                <span className="absolute right-2 flex size-3.5 items-center justify-center">
                  <CheckIcon className="size-4" />
                </span>
              )}
            </div>
            {hasChildren && isExpanded && (
              <TreeSelectBranch
                items={item.children!}
                level={level + 1}
                selectedValue={selectedValue}
                onSelect={onSelect}
                expandedIds={expandedIds}
                onToggleExpand={onToggleExpand}
              />
            )}
          </div>
        );
      })}
    </>
  );
}

export function TreeSelect<T extends TreeSelectNode<T>>({
  items,
  value,
  onValueChange,
  placeholder = "选择...",
  allowClear = false,
  clearLabel = "清空选择",
  disabled,
  className,
  size = "default",
  emptyMessage = "暂无数据",
}: TreeSelectProps<T>) {
  const contentId = React.useId();
  const [open, setOpen] = React.useState(false);
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set());

  const selectedItem = findItemByValue(items, value);

  const handleSelect = (selectedValue: string, node: T) => {
    onValueChange(selectedValue, node);
    setOpen(false);
  };

  const handleClear = () => {
    onValueChange("", null);
    setOpen(false);
  };

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  return (
    <Popover
      open={open}
      onOpenChange={(nextOpen) => {
        if (!disabled) {
          setOpen(nextOpen);
        }
      }}
    >
      <PopoverTrigger asChild>
        <button
          type="button"
          role="combobox"
          aria-controls={contentId}
          aria-expanded={open}
          aria-disabled={disabled}
          data-placeholder={selectedItem ? undefined : ""}
          data-size={size}
          disabled={disabled}
          className={cn(
            "border-input data-[placeholder]:text-muted-foreground [&_svg:not([class*='text-'])]:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 dark:hover:bg-input/50 flex w-fit items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm whitespace-nowrap shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 data-[size=default]:h-9 data-[size=sm]:h-8 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
            "w-full",
            className,
          )}
        >
          <span className="line-clamp-1 flex items-center gap-2">
            {selectedItem?.label || placeholder}
          </span>
          <ChevronDown className="size-4 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent
        id={contentId}
        align="start"
        sideOffset={4}
        className="max-h-80 overflow-y-auto p-1"
        style={{ width: "var(--radix-popover-trigger-width)" }}
      >
        {allowClear && (
          <>
            <div
              role="button"
              tabIndex={0}
              aria-pressed={!value}
              className={cn(
                "relative flex items-center gap-1 rounded-sm py-1.5 pr-8 pl-2 text-sm cursor-pointer",
                "hover:bg-accent hover:text-accent-foreground",
                "outline-none focus-visible:bg-accent focus-visible:text-accent-foreground",
                !value && "bg-accent text-accent-foreground",
              )}
              onClick={handleClear}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  handleClear();
                }
              }}
            >
              <span className="truncate">{clearLabel}</span>
              {!value && (
                <span className="absolute right-2 flex size-3.5 items-center justify-center">
                  <CheckIcon className="size-4" />
                </span>
              )}
            </div>
            {items.length > 0 && <div className="bg-border my-1 h-px" />}
          </>
        )}
        {items.length === 0 ? (
          <div className="py-6 text-center text-sm text-muted-foreground">{emptyMessage}</div>
        ) : (
          <div role="tree" aria-label={placeholder}>
            <TreeSelectBranch
              items={items}
              level={0}
              selectedValue={value}
              onSelect={handleSelect}
              expandedIds={expandedIds}
              onToggleExpand={toggleExpand}
            />
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}

function findItemByValue<T extends TreeSelectNode<T>>(
  items: TreeSelectItem<T>[],
  value?: string,
): TreeSelectItem<T> | undefined {
  if (!value) return undefined;
  for (const item of items) {
    if (item.value === value) return item;
    if (item.children) {
      const found = findItemByValue(item.children, value);
      if (found) return found;
    }
  }
  return undefined;
}
