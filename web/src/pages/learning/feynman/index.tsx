import { useNavigate } from "@tanstack/react-router";
import { BotIcon, CornerDownLeftIcon, Loader2Icon, PlayIcon, SendIcon } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

import { coachChatStream, type LearningCoachAction } from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Textarea } from "@/components/ui/textarea";

type CoachMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
};

export function FeynmanExercisePage() {
  const navigate = useNavigate();
  const bottomRef = useRef<HTMLDivElement>(null);
  const [messages, setMessages] = useState<CoachMessage[]>([]);
  const [input, setInput] = useState("继续学习");
  const [streaming, setStreaming] = useState(false);
  const [action, setAction] = useState<LearningCoachAction | null>(null);

  const scrollToBottom = () => {
    window.requestAnimationFrame(() => bottomRef.current?.scrollIntoView({ behavior: "smooth" }));
  };

  const send = async (rawMessage: string) => {
    const message = rawMessage.trim() || "继续学习";
    if (streaming) return;

    const assistantId = `assistant-${Date.now()}`;
    setAction(null);
    setMessages((prev) => [
      ...prev,
      { id: `user-${Date.now()}`, role: "user", content: message },
      { id: assistantId, role: "assistant", content: "" },
    ]);
    setStreaming(true);
    scrollToBottom();

    await coachChatStream(
      message,
      (event) => {
        if ((event.type === "stream_chunk" || event.type === "message") && event.content) {
          const visibleContent = event.content.replace(/<ACTION>[\s\S]*?<\/ACTION>/g, "").trimEnd();
          if (!visibleContent) return;
          setMessages((prev) =>
            prev.map((item) =>
              item.id === assistantId ? { ...item, content: item.content + visibleContent } : item,
            ),
          );
          scrollToBottom();
        } else if (event.type === "action" && event.action) {
          setAction(event.action);
        } else if (event.type === "error") {
          toast.error(event.content || "学习 agent 出错");
        }
      },
      () => setStreaming(false),
      (error) => {
        setStreaming(false);
        toast.error(error.message);
      },
    );
  };

  const handleSubmit = () => {
    const message = input.trim();
    if (!message || streaming) return;
    setInput("");
    void send(message);
  };

  const enterPractice = () => {
    if (action?.type !== "navigate_to_practice" || !action.objective_id) return;
    navigate({
      to: "/learn/feynman-practice/$objectiveId",
      params: { objectiveId: action.objective_id },
    });
  };

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-hidden p-6">
      <div className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-2xl font-bold">费曼学习 Agent</h1>
          <Badge variant="secondary">真实上下文</Badge>
        </div>
        <p className="max-w-3xl text-sm leading-6 text-muted-foreground">
          直接说继续学习。Agent 会查询 Wiki
          文件夹、文档、学习记录和用户画像，再决定下一步进入哪一个小节。
        </p>
      </div>

      <Card className="min-h-0 flex-1 gap-0 overflow-hidden rounded-lg py-0">
        <CardHeader className="border-b px-4 py-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <BotIcon className="size-4" />
            学习调度
          </CardTitle>
        </CardHeader>
        <CardContent className="flex min-h-0 flex-1 flex-col gap-0 p-0">
          <ScrollArea className="min-h-0 flex-1">
            <div className="flex min-h-[420px] flex-col gap-4 p-4">
              {messages.length === 0 ? (
                <div className="flex flex-1 flex-col items-center justify-center gap-4 text-center">
                  <div className="rounded-full border bg-muted/40 p-3">
                    <BotIcon className="size-6 text-muted-foreground" />
                  </div>
                  <div className="space-y-1">
                    <div className="text-base font-semibold">让 Agent 接管下一步</div>
                    <div className="max-w-md text-sm leading-6 text-muted-foreground">
                      它会先读当前资料结构和学习状态，再告诉你该继续、复习，还是先补资料。
                    </div>
                  </div>
                  <Button onClick={() => void send("继续学习")} disabled={streaming}>
                    <PlayIcon className="size-4" />
                    继续学习
                  </Button>
                </div>
              ) : (
                messages.map((message) => (
                  <div
                    key={message.id}
                    className={message.role === "user" ? "flex justify-end" : "flex justify-start"}
                  >
                    <div
                      className={
                        message.role === "user"
                          ? "max-w-[78%] whitespace-pre-wrap rounded-lg bg-primary px-3 py-2 text-sm leading-6 text-primary-foreground"
                          : "max-w-[78%] whitespace-pre-wrap rounded-lg border bg-muted/40 px-3 py-2 text-sm leading-6"
                      }
                    >
                      {message.content || (streaming ? "正在查询上下文..." : "")}
                    </div>
                  </div>
                ))
              )}
              {action?.type === "navigate_to_practice" && action.objective_id ? (
                <div className="flex justify-start">
                  <Button onClick={enterPractice}>
                    <CornerDownLeftIcon className="size-4" />
                    {action.label || "进入练习"}
                  </Button>
                </div>
              ) : null}
              <div ref={bottomRef} />
            </div>
          </ScrollArea>

          <div className="border-t p-4">
            <div className="flex gap-2">
              <Textarea
                className="max-h-32 min-h-11 resize-none"
                value={input}
                placeholder="继续学习 / 复习薄弱点 / 今天从哪里开始..."
                onChange={(event) => setInput(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" && !event.shiftKey) {
                    event.preventDefault();
                    handleSubmit();
                  }
                }}
                disabled={streaming}
              />
              <Button
                className="h-11 shrink-0"
                onClick={handleSubmit}
                disabled={streaming || !input.trim()}
              >
                {streaming ? (
                  <Loader2Icon className="size-4 animate-spin" />
                ) : (
                  <SendIcon className="size-4" />
                )}
                发送
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
