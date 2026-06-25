import { useCallback } from "react";

import { ChatPanel } from "./_components/chat-panel";
import { useAgentChat } from "./_hooks/agent-chat";

export function AgentPage() {
  const { messages, loading, sendMessage } = useAgentChat();

  const handleSend = useCallback(
    async (text: string) => {
      await sendMessage(text);
    },
    [sendMessage],
  );

  return (
    <div className="flex h-screen overflow-hidden">
      {/* 右侧聊天面板 */}
      <div className="flex-1">
        <ChatPanel messages={messages} loading={loading} onSend={handleSend} />
      </div>
    </div>
  );
}
