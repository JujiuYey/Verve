import logo from "@/assets/logo.svg";
import { SidebarHeader } from "@/components/ui/sidebar";

export function Logo() {
  return (
    <SidebarHeader>
      <div className="bg-sidebar-accent ring-sidebar-border/60 flex items-center gap-3 rounded-lg p-2 ring-1 shadow-[0_4px_14px_rgb(15_23_42/0.06),0_1px_2px_rgb(15_23_42/0.04)] group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-0 group-data-[collapsible=icon]:shadow-none group-data-[collapsible=icon]:ring-0">
        <div className="focus-visible:ring-sidebar-ring rounded-md focus-visible:ring-2 focus-visible:outline-none">
          <img src={logo} alt="SAG Wiki" className="size-9 rounded-md object-contain" />
        </div>
        <div className="flex flex-1 items-center gap-2 group-data-[collapsible=icon]:hidden">
          <div className="grid flex-1 text-left leading-tight">
            <span className="truncate text-[15px] font-semibold tracking-tight">SAG Wiki</span>
            <span className="text-muted-foreground truncate text-xs">智能知识运营中枢</span>
          </div>
        </div>
      </div>
    </SidebarHeader>
  );
}
