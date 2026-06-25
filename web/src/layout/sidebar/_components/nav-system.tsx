import {
  IconBuilding,
  IconListDetails,
  IconShield,
  IconSparkles,
  IconUsers,
} from "@tabler/icons-react";
import { Link } from "@tanstack/react-router";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

const data = [
  {
    title: "部门管理",
    url: "/system/department",
    icon: IconBuilding,
  },
  {
    title: "角色管理",
    url: "/system/role",
    icon: IconShield,
  },
  {
    title: "用户管理",
    url: "/system/user",
    icon: IconUsers,
  },
  {
    title: "队列监控",
    url: "/system/queue",
    icon: IconListDetails,
  },
  {
    title: "系统助手",
    url: "/system/agent",
    icon: IconSparkles,
  },
];

export function NavSystem() {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>系统管理</SidebarGroupLabel>
      <SidebarGroupContent className="flex flex-col gap-2">
        <SidebarMenu>
          {data.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton tooltip={item.title} asChild>
                <Link to={item.url}>
                  {item.icon && <item.icon />}
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
