import {
  AgentIcon,
  ComposeIcon,
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
  { title: "费曼练习", url: "/learn/feynman", icon: AgentIcon },
  { title: "日志", url: "/learn/journal", icon: DocumentIcon },
  { title: "我的画像", url: "/learn/profile", icon: DashboardIcon },
];

export const knowledgeNav: NavItem[] = [
  { title: "文档管理", url: "/wiki/folders", icon: FolderIcon },
  { title: "新建文档", url: "/wiki/tiptap-editor", icon: ComposeIcon },
];

export const systemNav: NavItem[] = [
  { title: "用户管理", url: "/system/user", icon: UserManageIcon },
  { title: "模型配置", url: "/system/model-config", icon: ModelConfigIcon },
];
