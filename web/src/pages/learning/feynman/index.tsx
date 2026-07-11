import { useNavigate } from "@tanstack/react-router";
import { getRouteApi } from "@tanstack/react-router";
import type { ChatStatus } from "ai";
import { useState } from "react";
import { toast } from "sonner";

import { coachChatStream, type LearningCoachAction } from "@/api/learning";
import { PromptInputProvider } from "@/components/ai-elements/prompt-input";

import { CoachWorkspace, type CoachMessage, type ToolEvent } from "./_components/coach-workspace";

const STRIP_ACTIONS = /<ACTION>[\s\S]*?<\/ACTION>/g;
const routeApi = getRouteApi("/_layout/learn/feynman");

function stripAction(content: string): string {
  return content.replace(STRIP_ACTIONS, "").trimEnd();
}

export function FeynmanExercisePage() {
  const navigate = useNavigate();
  const search = routeApi.useSearch();
  const [messages, setMessages] = useState<CoachMessage[]>([]);
  const [status, setStatus] = useState<ChatStatus>("ready");
  const [action, setAction] = useState<LearningCoachAction | null>(null);

  const updateAssistant = (assistantId: string, patch: (m: CoachMessage) => CoachMessage) => {
    setMessages((prev) => prev.map((item) => (item.id === assistantId ? patch(item) : item)));
  };

  const send = async (rawMessage: string) => {
    const message = rawMessage.trim() || "继续学习";
    if (status === "submitted" || status === "streaming") return;

    const assistantId = `assistant-${Date.now()}`;
    setAction(null);
    setMessages((prev) => [
      ...prev,
      { id: `user-${Date.now()}`, role: "user", content: message },
      { id: assistantId, role: "assistant", content: "" },
    ]);
    setStatus("submitted");

    await coachChatStream(
      message,
      (event) => {
        switch (event.type) {
          case "reasoning": {
            if (!event.content) return;
            setStatus("streaming");
            updateAssistant(assistantId, (m) => ({
              ...m,
              reasoning: (m.reasoning || "") + event.content,
            }));
            return;
          }
          case "stream_chunk":
          case "message": {
            if (!event.content) return;
            const cleaned = stripAction(event.content);
            if (!cleaned) return;
            setStatus("streaming");
            updateAssistant(assistantId, (m) => ({ ...m, content: m.content + cleaned }));
            return;
          }
          case "tool_call": {
            if (!event.id || !event.name) return;
            const next: ToolEvent = {
              id: event.id,
              name: event.name,
              arguments: event.arguments,
              state: "input-available",
            };
            updateAssistant(assistantId, (m) => ({
              ...m,
              toolEvents: [...(m.toolEvents || []), next],
            }));
            return;
          }
          case "tool_result": {
            if (!event.tool_call_id) return;
            updateAssistant(assistantId, (m) => ({
              ...m,
              toolEvents: (m.toolEvents || []).map((t) =>
                t.id === event.tool_call_id
                  ? {
                      ...t,
                      output: event.content || t.output,
                      state: "output-available",
                    }
                  : t,
              ),
            }));
            return;
          }
          case "action": {
            if (event.action) setAction(event.action);
            return;
          }
          case "error": {
            setStatus("error");
            toast.error(event.content || "学习 agent 出错");
            return;
          }
        }
      },
      () => setStatus("ready"),
      (error) => {
        setStatus("error");
        toast.error(error.message);
      },
      {
        agent_instance_id: search.agentInstanceId,
        root_folder_id: search.rootFolderId,
      },
    );
  };

  const enterPractice = () => {
    if (action?.type !== "navigate_to_practice" || !action.document_id) return;
    navigate({
      to: "/learn/feynman-practice/$documentId",
      params: { documentId: action.document_id },
    });
  };

  return (
    <PromptInputProvider initialInput="继续学习">
      <CoachWorkspace
        action={action}
        agentName={search.rootFolderName ? `${search.rootFolderName} 学习 Agent` : undefined}
        messages={messages}
        onEnterPractice={enterPractice}
        onSend={(message) => void send(message)}
        status={status}
      />
    </PromptInputProvider>
  );
}
