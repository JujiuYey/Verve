import { Button } from "@/components/ui/button";

interface BreadcrumbNavProps {
  breadcrumb: { id: string; name: string }[];
  onNavigate: (index: number) => void;
}

export function BreadcrumbNav({ breadcrumb, onNavigate }: BreadcrumbNavProps) {
  return (
    <div className="px-3 pb-2 flex items-center gap-1 text-sm">
      {breadcrumb.length === 0 ? (
        <span className="text-muted-foreground">组织架构</span>
      ) : (
        <>
          <Button variant="ghost" className="h-auto px-1.5 py-0.5" onClick={() => onNavigate(-1)}>
            组织架构
          </Button>
          {breadcrumb.map((crumb, index) => (
            <span key={crumb.id} className="flex items-center gap-1">
              <span className="text-muted-foreground">&gt;</span>
              {index === breadcrumb.length - 1 ? (
                <span className="text-muted-foreground">{crumb.name}</span>
              ) : (
                <Button variant="link" className="h-auto p-0" onClick={() => onNavigate(index)}>
                  {crumb.name}
                </Button>
              )}
            </span>
          ))}
        </>
      )}
    </div>
  );
}
