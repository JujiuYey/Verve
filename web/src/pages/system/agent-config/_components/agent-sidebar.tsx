import { cn } from "@/lib/utils";

import {
  AGENTS,
  getAgentStatus,
  type AgentDefinition,
  type SceneConfigState,
} from "../agent-definitions";
import { StatusBadge } from "./status-badge";

interface AgentSidebarProps {
  selectedAgent?: AgentDefinition;
  configsByScene: Map<string, SceneConfigState>;
  onSelectAgent: (agentKey: string) => void;
}

export function AgentSidebar({ selectedAgent, configsByScene, onSelectAgent }: AgentSidebarProps) {
  return (
    <aside className="flex min-h-0 flex-col border-r bg-muted/20">
      <nav className="flex min-h-0 flex-1 flex-col gap-1 overflow-auto p-3">
        {AGENTS.map((agent) => {
          const status = getAgentStatus(agent, configsByScene);
          const Icon = agent.icon;
          const selected = selectedAgent?.key === agent.key;
          return (
            <button
              key={agent.key}
              type="button"
              className={cn(
                "flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
                selected ? "bg-background shadow-xs ring-1 ring-border" : "hover:bg-background/70",
              )}
              onClick={() => onSelectAgent(agent.key)}
            >
              <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-background ring-1 ring-border">
                <Icon />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-medium">{agent.shortName}</span>
                <span className="block truncate font-mono text-xs text-muted-foreground">
                  {agent.runtime}
                </span>
              </span>
              <StatusBadge status={status} />
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
