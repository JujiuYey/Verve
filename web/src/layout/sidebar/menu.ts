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
  QueueIcon,
  RoleIcon,
  UserManageIcon,
  type SidebarNavIcon,
} from "./nav-icons";

export type NavItem = {
  title: string;
  url: string;
  icon: SidebarNavIcon;
};

export const aiNav: NavItem[] = [
  { title: "仪表盘", url: "/", icon: DashboardIcon },
  { title: "RAG 对话", url: "/rag-chat", icon: ChatIcon },
  { title: "模型配置", url: "/model-config", icon: ModelConfigIcon },
  { title: "向量数据库", url: "/collection", icon: CollectionIcon },
];

export const knowledgeNav: NavItem[] = [
  { title: "文档管理", url: "/wiki/folders", icon: FolderIcon },
  { title: "知识库", url: "/wiki/documents", icon: DocumentIcon },
  { title: "新建文档", url: "/wiki/tiptap-editor", icon: ComposeIcon },
];

export const systemNav: NavItem[] = [
  { title: "部门管理", url: "/system/department", icon: DepartmentIcon },
  { title: "角色管理", url: "/system/role", icon: RoleIcon },
  { title: "用户管理", url: "/system/user", icon: UserManageIcon },
  { title: "队列监控", url: "/system/queue", icon: QueueIcon },
  { title: "系统助手", url: "/system/agent", icon: AgentIcon },
];
