import { Link } from "@tanstack/react-router";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

import { knowledgeNav, learnNav, systemNav, type NavItem } from "./menu";

const activeClass =
  "bg-sidebar-accent text-sidebar-foreground font-medium shadow-[0_1px_3px_rgb(15_23_42_/_0.08),0_1px_2px_rgb(15_23_42_/_0.04)]";

function NavGroup({ label, items }: { label: string; items: NavItem[] }) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>{label}</SidebarGroupLabel>
      <SidebarGroupContent className="flex flex-col gap-2">
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton tooltip={item.title} asChild>
                <Link
                  className="group/nav-link"
                  to={item.url}
                  activeProps={{ className: activeClass }}
                >
                  <item.icon />
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

export function NavSystem() {
  return (
    <>
      <NavGroup label="学习" items={learnNav} />
      <NavGroup label="知识库" items={knowledgeNav} />
      <NavGroup label="系统管理" items={systemNav} />
    </>
  );
}
