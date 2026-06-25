import { IconMessageCirclePlus, IconSearch } from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";

import { ChatPanel } from "./_components/chat-panel";
import { FolderList } from "./_components/folder-list";
import { SessionList } from "./_components/session-list";

interface RagSession {
  id: string;
  title?: string;
  folder_id: string;
  document_id?: string;
  message_count: number;
  total_tokens?: number;
  created_at: string;
  updated_at: string;
}

function generateId(): string {
  return `mock-rag-session-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

// Mock 会话列表
const mockSessions: RagSession[] = [
  {
    id: generateId(),
    title: "RAG 知识问答",
    folder_id: "folder-1",
    message_count: 5,
    total_tokens: 1200,
    created_at: new Date(Date.now() - 3600000 * 24).toISOString(),
    updated_at: new Date(Date.now() - 3600000 * 2).toISOString(),
  },
  {
    id: generateId(),
    title: "文档检索对话",
    folder_id: "folder-1",
    message_count: 3,
    total_tokens: 800,
    created_at: new Date(Date.now() - 3600000 * 48).toISOString(),
    updated_at: new Date(Date.now() - 3600000 * 24).toISOString(),
  },
];

export function RagChatPage() {
  const [activeFolder, setActiveFolder] = useState<string>();
  const [selectedSession, setSelectedSession] = useState<string>();
  const [sessions, setSessions] = useState<RagSession[]>([]);
  const [searchValue, setSearchValue] = useState("");

  const fetchSessions = useCallback(async () => {
    try {
      // 模拟网络延迟
      await new Promise((resolve) => setTimeout(resolve, 300 + Math.random() * 200));
      setSessions([...mockSessions]);
    } catch (error) {
      console.error("获取会话列表失败:", error);
    }
  }, []);

  // Load sessions on mount
  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  // When folder changes, reset selected session and reload sessions
  const handleFolderChange = (id: string) => {
    setActiveFolder(id);
    setSelectedSession(undefined);
    fetchSessions();
  };

  // Filter sessions by selected folder and search
  const filteredSessions = sessions.filter((s) => {
    if (activeFolder && s.folder_id !== activeFolder) return false;
    if (searchValue) {
      const keyword = searchValue.toLowerCase();
      return (s.title || "").toLowerCase().includes(keyword);
    }
    return true;
  });

  // Auto-select first session when filtered list changes
  useEffect(() => {
    if (filteredSessions.length > 0 && !selectedSession) {
      setSelectedSession(filteredSessions[0]!.id);
    }
  }, [filteredSessions.length, activeFolder]);

  const handleCreateNewSession = () => {
    setSelectedSession(undefined);
  };

  const handleSessionCreated = (sessionId: string) => {
    fetchSessions();
    setSelectedSession(sessionId);
  };

  const handleSessionDeleted = () => {
    setSelectedSession(undefined);
    fetchSessions();
  };

  return (
    <div className="flex h-[calc(100vh-1rem)] overflow-hidden">
      {/* Left panel: Knowledge Bases */}
      <div className="w-60 shrink-0 border-r overflow-hidden">
        <FolderList value={activeFolder} onChange={handleFolderChange} />
      </div>

      {/* Middle panel: Session List */}
      <div className="w-[320px] shrink-0 border-r flex flex-col overflow-hidden">
        <div className="px-4 py-2 flex items-center justify-between">
          <h1 className="text-xl font-bold">对话历史</h1>
          <Button size="sm" onClick={handleCreateNewSession}>
            <IconMessageCirclePlus className="mr-1 h-4 w-4" />
            新建对话
          </Button>
        </div>
        <Separator />
        <div className="p-4 bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
          <div className="relative">
            <IconSearch className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索"
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>
        <SessionList
          items={filteredSessions}
          selectedId={selectedSession}
          onSelect={setSelectedSession}
          onDeleted={handleSessionDeleted}
          onTitleUpdated={fetchSessions}
        />
      </div>

      {/* Right panel: Chat */}
      <div className="flex-1 min-w-0 overflow-hidden">
        <ChatPanel
          sessionId={selectedSession}
          folderId={activeFolder}
          onSessionCreated={handleSessionCreated}
        />
      </div>
    </div>
  );
}
