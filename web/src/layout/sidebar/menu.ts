import {
  AgentIcon,
  ChatIcon,
  DashboardIcon,
  DocumentIcon,
  FolderIcon,
  ModelConfigIcon,
  type SidebarNavIcon,
} from "./nav-icons";

export type NavItem = {
  title: string;
  url: string;
  icon: SidebarNavIcon;
};

export const learnNav: NavItem[] = [
  { title: "知识库", url: "/wiki", icon: FolderIcon },
  { title: "知识问答", url: "/learn/ask", icon: ChatIcon },
  { title: "日志", url: "/learn/journal", icon: DocumentIcon },
  { title: "学习记忆", url: "/learn/profile", icon: DashboardIcon },
];

export const systemNav: NavItem[] = [
  { title: "模型配置", url: "/system/model-config", icon: ModelConfigIcon },
  { title: "Agent 配置", url: "/system/agent-config", icon: AgentIcon },
];
