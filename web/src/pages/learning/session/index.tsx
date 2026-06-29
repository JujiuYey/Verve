import { useNavigate, useParams } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  sessionChatStream,
  useCompleteSession,
  useSessionDetail,
  useSubmitExercise,
  type ExerciseResult,
} from "@/api/learning";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

const EXERCISE_TYPES = [
  { value: "explain", label: "解释" },
  { value: "code_snippet", label: "代码" },
  { value: "paste_output", label: "运行结果" },
];

interface UIMessage {
  id: string;
  role: string;
  content: string;
}

export function SessionPage() {
  const { sessionId } = useParams({ from: "/_layout/learn/session/$sessionId" });
  const navigate = useNavigate();

  const { data: detail } = useSessionDetail(sessionId);
  const submitExercise = useSubmitExercise(sessionId);
  const completeSession = useCompleteSession(sessionId);

  const [messages, setMessages] = useState<UIMessage[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);

  const [exType, setExType] = useState("explain");
  const [answer, setAnswer] = useState("");
  const [verdict, setVerdict] = useState<ExerciseResult | null>(null);

  const bottomRef = useRef<HTMLDivElement>(null);

  // 载入历史消息
  useEffect(() => {
    if (!detail) return;
    setMessages(
      (detail.messages || []).map((m) => ({ id: m.id, role: m.role, content: m.content })),
    );
  }, [detail]);

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const send = async (message: string) => {
    if (streaming) return;
    if (message.trim()) {
      setMessages((prev) => [...prev, { id: `u-${Date.now()}`, role: "user", content: message }]);
    }
    const assistantId = `a-${Date.now()}`;
    setMessages((prev) => [...prev, { id: assistantId, role: "assistant", content: "" }]);
    setStreaming(true);

    await sessionChatStream(
      sessionId,
      message,
      (event) => {
        if ((event.type === "stream_chunk" || event.type === "message") && event.content) {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === assistantId ? { ...m, content: m.content + event.content } : m,
            ),
          );
          scrollToBottom();
        } else if (event.type === "error") {
          toast.error(event.content || "对话出错");
        }
      },
      () => setStreaming(false),
      (err) => {
        setStreaming(false);
        toast.error(err.message);
      },
    );
  };

  const handleSend = () => {
    const t = input.trim();
    if (!t) return;
    setInput("");
    void send(t);
  };

  const handleSubmitExercise = async () => {
    if (!answer.trim()) return;
    const lastTutor = [...messages].reverse().find((m) => m.role === "assistant");
    const res = await submitExercise.mutateAsync({
      type: exType,
      prompt: lastTutor?.content?.slice(0, 500) || "",
      user_answer: answer,
    });
    setVerdict(res);
    setAnswer("");
  };

  const handleComplete = async () => {
    const res = await completeSession.mutateAsync();
    toast.success(res.summary || "本节完成");
    navigate({
      to: "/learn/goal/$goalId",
      params: { goalId: detail?.session.goal_id || "" },
    });
  };

  return (
    <div className="flex h-full flex-col p-6">
      <div className="mb-3 flex items-center justify-between">
        <h1 className="text-lg font-semibold">陪练</h1>
        <Button
          variant="outline"
          size="sm"
          onClick={handleComplete}
          disabled={completeSession.isPending}
        >
          结束本节
        </Button>
      </div>

      {/* 对话区 */}
      <div className="flex-1 overflow-auto rounded-md border p-4">
        {messages.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <Button onClick={() => void send("")} disabled={streaming}>
              开始这一节
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {messages.map((m) => (
              <div key={m.id} className={m.role === "user" ? "text-right" : "text-left"}>
                <div
                  className={`inline-block max-w-[80%] whitespace-pre-wrap rounded-lg px-3 py-2 text-sm ${
                    m.role === "user" ? "bg-primary text-primary-foreground" : "bg-muted"
                  }`}
                >
                  {m.content || (streaming ? "…" : "")}
                </div>
              </div>
            ))}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      {/* 输入 */}
      <div className="mt-3 flex gap-2">
        <Input
          placeholder="回答 / 提问…"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleSend();
          }}
          disabled={streaming}
        />
        <Button onClick={handleSend} disabled={streaming || !input.trim()}>
          发送
        </Button>
      </div>

      {/* 验证区 */}
      <Card className="mt-4 p-4">
        <div className="mb-2 flex flex-wrap items-center gap-2">
          <span className="text-sm font-medium">提交作答验证</span>
          <div className="flex gap-1">
            {EXERCISE_TYPES.map((t) => (
              <Badge
                key={t.value}
                variant={exType === t.value ? "default" : "outline"}
                className="cursor-pointer"
                onClick={() => setExType(t.value)}
              >
                {t.label}
              </Badge>
            ))}
          </div>
        </div>
        <Textarea
          placeholder="把你的解释 / 代码 / 运行结果贴在这里…"
          value={answer}
          onChange={(e) => setAnswer(e.target.value)}
          rows={4}
        />
        <div className="mt-2 flex justify-end">
          <Button
            onClick={handleSubmitExercise}
            disabled={submitExercise.isPending || !answer.trim()}
          >
            {submitExercise.isPending ? "判定中…" : "提交验证"}
          </Button>
        </div>
        {verdict ? (
          <div className="mt-3 rounded-md border p-3 text-sm">
            <div className="mb-1 flex items-center gap-2">
              <Badge
                variant={
                  verdict.verdict === "pass"
                    ? "default"
                    : verdict.verdict === "fail"
                      ? "destructive"
                      : "secondary"
                }
              >
                {verdict.verdict}
              </Badge>
              <span className="text-muted-foreground">掌握:{verdict.mastery_after}</span>
            </div>
            <div className="whitespace-pre-wrap">{verdict.feedback}</div>
          </div>
        ) : null}
      </Card>
    </div>
  );
}
