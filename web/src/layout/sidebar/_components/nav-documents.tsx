import { IconDatabase, IconFileDescription, IconFilePencil } from "@tabler/icons-react";
import { Link } from "@tanstack/react-router";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

const data = [
  {
    name: "文档管理",
    url: "/wiki/folders",
    icon: IconDatabase,
  },
  {
    name: "知识库",
    url: "/wiki/documents",
    icon: IconFileDescription,
  },
  {
    name: "新建文档",
    url: "/wiki/tiptap-editor",
    icon: IconFilePencil,
  },
];

export function NavDocuments() {
  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel>知识中心</SidebarGroupLabel>
      <SidebarMenu>
        {data.map((item) => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton asChild>
              <Link to={item.url}>
                <item.icon />
                <span>{item.name}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
