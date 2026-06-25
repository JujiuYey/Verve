import type { ReactNode, SVGProps } from "react";

import { cn } from "@/lib/utils";

export type SidebarNavIcon = (props: SVGProps<SVGSVGElement>) => ReactNode;

type SidebarIconProps = SVGProps<SVGSVGElement> & {
  children: ReactNode;
};

function SidebarIcon({ className, children, ...props }: SidebarIconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("sidebar-nav-icon !size-5 group-data-[collapsible=icon]:!size-6", className)}
      fill="none"
      focusable="false"
      viewBox="0 0 24 24"
      {...props}
    >
      {children}
    </svg>
  );
}

export function DashboardIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <rect data-tone="blue" x="3" y="4" width="7" height="7" rx="2" fill="currentColor" />
      <rect data-tone="cyan" x="14" y="4" width="7" height="7" rx="2" fill="currentColor" />
      <rect data-tone="green" x="3" y="15" width="7" height="7" rx="2" fill="currentColor" />
      <path
        data-tone="amber"
        d="M16.4 15h2.2c1.33 0 2.4 1.07 2.4 2.4v2.2c0 1.33-1.07 2.4-2.4 2.4h-2.2A2.4 2.4 0 0 1 14 19.6v-2.2c0-1.33 1.07-2.4 2.4-2.4Z"
        fill="currentColor"
      />
    </SidebarIcon>
  );
}

export function ChatIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M6.5 4h11A2.5 2.5 0 0 1 20 6.5v7A2.5 2.5 0 0 1 17.5 16H12l-4.25 3.25c-.65.5-1.55.04-1.55-.78V16H6.5A2.5 2.5 0 0 1 4 13.5v-7A2.5 2.5 0 0 1 6.5 4Z"
        fill="currentColor"
      />
      <circle data-tone="white" cx="9" cy="10" r="1.2" fill="currentColor" />
      <circle data-tone="white" cx="12" cy="10" r="1.2" fill="currentColor" />
      <circle data-tone="cyan" cx="15" cy="10" r="1.2" fill="currentColor" />
    </SidebarIcon>
  );
}

export function ModelConfigIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M13.1 2.8 4.6 13.2c-.75.92-.1 2.3 1.1 2.3h5.1l-1 5.1c-.28 1.43 1.51 2.31 2.47 1.2l8.93-10.35c.8-.92.14-2.36-1.08-2.36H14.6l1.02-4.93c.3-1.43-1.58-2.45-2.52-1.36Z"
        fill="currentColor"
      />
      <path
        data-tone="amber"
        d="m11.2 9.3-2.1 3.1h4.2l-.55 3.1 3.55-4.45h-4.45l.6-1.75h-1.25Z"
        fill="currentColor"
      />
    </SidebarIcon>
  );
}

export function CollectionIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <ellipse data-tone="blue" cx="12" cy="6" rx="7" ry="3" fill="currentColor" />
      <path data-tone="cyan" d="M5 6v5c0 1.66 3.13 3 7 3s7-1.34 7-3V6" fill="currentColor" />
      <path data-tone="green" d="M5 11v5c0 1.66 3.13 3 7 3s7-1.34 7-3v-5" fill="currentColor" />
      <path
        data-tone="white"
        d="M9 9.5h6M9 14.5h6"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.8"
      />
    </SidebarIcon>
  );
}

export function FolderIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="amber"
        d="M3.5 7.5A2.5 2.5 0 0 1 6 5h3.2c.58 0 1.13.2 1.56.56l1.05.88c.43.36.98.56 1.54.56H18A2.5 2.5 0 0 1 20.5 9.5V17A3 3 0 0 1 17.5 20h-11A3 3 0 0 1 3.5 17V7.5Z"
        fill="currentColor"
      />
      <path data-tone="blue" d="M3.5 10.5h17V17a3 3 0 0 1-3 3h-11a3 3 0 0 1-3-3v-6.5Z" fill="currentColor" />
      <path data-tone="white" d="M7.5 14h5" stroke="currentColor" strokeLinecap="round" strokeWidth="1.8" />
    </SidebarIcon>
  );
}

export function DocumentIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M6 3h8.4L19 7.6V19a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z"
        fill="currentColor"
      />
      <path data-tone="cyan" d="M14 3v4a1 1 0 0 0 1 1h4" fill="currentColor" />
      <path
        data-tone="white"
        d="M8 11h7M8 14.5h7M8 18h4.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.8"
      />
    </SidebarIcon>
  );
}

export function ComposeIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M6.5 3h9A2.5 2.5 0 0 1 18 5.5V19a2 2 0 0 1-2 2H6.5A2.5 2.5 0 0 1 4 18.5v-13A2.5 2.5 0 0 1 6.5 3Z"
        fill="currentColor"
      />
      <path data-tone="cyan" d="M8 3v18" stroke="currentColor" strokeWidth="2" />
      <path
        data-tone="white"
        d="M11 8h3.8M11 11.5h3"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.7"
      />
      <path
        data-tone="amber"
        d="m13.2 17.6 4.85-4.85 2.2 2.2-4.85 4.85-2.75.55.55-2.75Z"
        fill="currentColor"
      />
    </SidebarIcon>
  );
}

export function DepartmentIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M4 9.2 11.4 4a1 1 0 0 1 1.2 0L20 9.2v1.3a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V9.2Z"
        fill="currentColor"
      />
      <rect data-tone="cyan" x="6" y="12" width="3" height="8" rx="1" fill="currentColor" />
      <rect data-tone="green" x="10.5" y="12" width="3" height="8" rx="1" fill="currentColor" />
      <rect data-tone="amber" x="15" y="12" width="3" height="8" rx="1" fill="currentColor" />
      <path data-tone="blue" d="M4.6 20h14.8" stroke="currentColor" strokeLinecap="round" strokeWidth="2" />
    </SidebarIcon>
  );
}

export function RoleIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M12 3.5 5.25 6v5.2c0 4.2 2.82 8.1 6.75 9.8 3.93-1.7 6.75-5.6 6.75-9.8V6L12 3.5Z"
        fill="currentColor"
      />
      <path
        data-tone="white"
        d="m9.4 12.2 1.8 1.8 3.5-3.8"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
      />
      <path data-tone="cyan" d="M12 6.2v2.1" stroke="currentColor" strokeLinecap="round" strokeWidth="1.8" />
    </SidebarIcon>
  );
}

export function UserManageIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <circle data-tone="blue" cx="9" cy="8" r="4" fill="currentColor" />
      <path
        data-tone="cyan"
        d="M3.5 18.1C4.3 14.95 6.35 13 9 13s4.7 1.95 5.5 5.1c.3 1.17-.6 2.25-1.8 2.25H5.3c-1.2 0-2.1-1.08-1.8-2.25Z"
        fill="currentColor"
      />
      <circle data-tone="amber" cx="16.8" cy="9" r="2.9" fill="currentColor" />
      <path
        data-tone="green"
        d="M14.1 19.1c.48-2.12 1.82-3.42 3.55-3.42 1.72 0 3.05 1.3 3.54 3.42.24 1.02-.54 1.95-1.58 1.95h-3.93c-1.04 0-1.81-.93-1.58-1.95Z"
        fill="currentColor"
      />
    </SidebarIcon>
  );
}

export function QueueIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <path
        data-tone="blue"
        d="M6 3h8.4L19 7.6V19a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z"
        fill="currentColor"
      />
      <path data-tone="cyan" d="M14 3v4a1 1 0 0 0 1 1h4" fill="currentColor" />
      <path data-tone="green" d="M8 12h8v2H8zM8 16h5v2H8z" fill="currentColor" />
      <circle data-tone="amber" cx="16.8" cy="16.9" r="2.2" fill="currentColor" />
    </SidebarIcon>
  );
}

export function AgentIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <SidebarIcon {...props}>
      <circle data-tone="blue" cx="12" cy="12" r="8" fill="currentColor" />
      <circle data-tone="white" cx="9.2" cy="11" r="1.2" fill="currentColor" />
      <circle data-tone="white" cx="14.8" cy="11" r="1.2" fill="currentColor" />
      <path
        data-tone="cyan"
        d="M8.8 14.8c.85 1 1.9 1.5 3.2 1.5 1.3 0 2.35-.5 3.2-1.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.9"
      />
      <path data-tone="amber" d="M12 4V2.5" stroke="currentColor" strokeLinecap="round" strokeWidth="2" />
    </SidebarIcon>
  );
}
