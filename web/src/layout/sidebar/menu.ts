import {
  AgentIcon,
  ChatIcon,
  CollectionIcon,
  ComposeIcon,
  DashboardIcon,
  DepartmentIcon,
  DocumentIcon,
  FolderIcon,
  ModelConfigIcon,
  RoleIcon,
  UserManageIcon,
  type SidebarNavIcon,
} from "./nav-icons";

export type NavItem = {
  title: string;
  url: string;
  icon: SidebarNavIcon;
};

export const learnNav: NavItem[] = [
  { title: "开始学习", url: "/", icon: ChatIcon },
  { title: "费曼练习", url: "/learn/feynman", icon: AgentIcon },
  { title: "日志", url: "/learn/journal", icon: DocumentIcon },
  { title: "我的画像", url: "/learn/profile", icon: DashboardIcon },
];

export const knowledgeNav: NavItem[] = [
  { title: "文档管理", url: "/wiki/folders", icon: FolderIcon },
  { title: "知识库", url: "/wiki/documents", icon: CollectionIcon },
  { title: "新建文档", url: "/wiki/tiptap-editor", icon: ComposeIcon },
];

export const configNav: NavItem[] = [
  { title: "模型配置", url: "/model-config", icon: ModelConfigIcon },
];

export const systemNav: NavItem[] = [
  { title: "部门管理", url: "/system/department", icon: DepartmentIcon },
  { title: "角色管理", url: "/system/role", icon: RoleIcon },
  { title: "用户管理", url: "/system/user", icon: UserManageIcon },
];
