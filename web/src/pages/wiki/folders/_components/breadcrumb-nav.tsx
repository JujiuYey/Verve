import { IconChevronRight, IconHome } from "@tabler/icons-react";
import { useCallback } from "react";

import { Button } from "@/components/ui/button";

interface BreadcrumbItem {
  id?: string;
  name: string;
}

interface BreadcrumbNavProps {
  items: BreadcrumbItem[];
  onNavigate: (folder: BreadcrumbItem | null, index: number) => void;
}

export function BreadcrumbNav({ items, onNavigate }: BreadcrumbNavProps) {
  const handleRootClick = useCallback(() => {
    onNavigate(null, -1);
  }, [onNavigate]);

  const handleItemClick = useCallback(
    (item: BreadcrumbItem, index: number) => {
      onNavigate(item, index);
    },
    [onNavigate],
  );

  return (
    <nav className="flex items-center gap-1 text-sm">
      <Button variant="ghost" size="sm" className="gap-1 px-2" onClick={handleRootClick}>
        <IconHome className="h-4 w-4" />
        根目录
      </Button>

      {items.map((item, index) => (
        <div key={item.id ?? "root"} className="flex items-center gap-1">
          <IconChevronRight className="h-4 w-4 text-muted-foreground" />
          <Button
            variant="ghost"
            size="sm"
            className={index === items.length - 1 ? "font-normal" : ""}
            onClick={() => handleItemClick(item, index)}
          >
            {item.name}
          </Button>
        </div>
      ))}
    </nav>
  );
}
