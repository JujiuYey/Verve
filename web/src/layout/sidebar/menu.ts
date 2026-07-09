import {
  AgentIcon,
  DashboardIcon,
  DocumentIcon,
  FolderIcon,
  ModelConfigIcon,
  UserManageIcon,
  type SidebarNavIcon,
} from "./nav-icons";

export type NavItem = {
  title: string;
  url: string;
  icon: SidebarNavIcon;
};

export const learnNav: NavItem[] = [
  { title: "知识库", url: "/wiki", icon: FolderIcon },
  { title: "费曼练习", url: "/learn/feynman", icon: AgentIcon },
  { title: "日志", url: "/learn/journal", icon: DocumentIcon },
  { title: "我的画像", url: "/learn/profile", icon: DashboardIcon },
];

export const systemNav: NavItem[] = [
  { title: "用户管理", url: "/system/user", icon: UserManageIcon },
  { title: "模型配置", url: "/system/model-config", icon: ModelConfigIcon },
  { title: "Agent 配置", url: "/system/agent-config", icon: AgentIcon },
];
